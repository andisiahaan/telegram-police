package handler

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"sync"

	"github.com/andisiahaan/telegram-police/internal/telegram"
)

// WebhookHandler handles incoming POST requests from Telegram.
type WebhookHandler struct {
	dispatcher *Dispatcher
	secret     string
	wg         *sync.WaitGroup
}

// NewWebhookHandler creates a WebhookHandler.
func NewWebhookHandler(dispatcher *Dispatcher, secret string, wg *sync.WaitGroup) *WebhookHandler {
	return &WebhookHandler{
		dispatcher: dispatcher,
		secret:     secret,
		wg:         wg,
	}
}

// ServeHTTP validates the request, decodes the update, responds 200 immediately,
// then dispatches the update in a goroutine so Telegram doesn't time out.
func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if h.secret != "" {
		received := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")
		if received != h.secret {
			slog.Warn("webhook: invalid secret token", "remote_addr", r.RemoteAddr)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("webhook: failed to read body", "error", err)
		w.WriteHeader(http.StatusOK) // always 200 so Telegram doesn't retry
		return
	}
	defer r.Body.Close()

	var update telegram.Update
	if err := json.Unmarshal(body, &update); err != nil {
		slog.Error("webhook: failed to parse update", "error", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Respond immediately so Telegram doesn't time out and retry.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		slog.Error("webhook: failed to write json response", "error", err)
	}

	// Dispatch asynchronously
	if h.wg != nil {
		h.wg.Add(1)
	}
	go func(u telegram.Update) {
		if h.wg != nil {
			defer h.wg.Done()
		}
		h.dispatcher.Dispatch(&u)
	}(update)
}
