package handler

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/andisiahaan/telegram-police/internal/telegram"
)

// WebhookHandler handles incoming POST requests from Telegram.
type WebhookHandler struct {
	dispatcher *Dispatcher
	secret     string
}

// NewWebhookHandler creates a WebhookHandler.
func NewWebhookHandler(dispatcher *Dispatcher, secret string) *WebhookHandler {
	return &WebhookHandler{
		dispatcher: dispatcher,
		secret:     secret,
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

	ctx := h.dispatcher.Dispatch(&update)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := map[string]interface{}{
		"status": "ok",
	}
	if ctx != nil {
		if ctx.Bypassed {
			resp["bypassed"] = true
			resp["message"] = ctx.BypassReason
		}
		if len(ctx.Violations) > 0 {
			resp["violations"] = ctx.Violations
		}
		if len(ctx.ExecutedActions) > 0 {
			resp["executed_actions"] = ctx.ExecutedActions
		}
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("webhook: failed to write json response", "error", err)
	}
}
