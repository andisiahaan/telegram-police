package admin

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/andisiahaan/telegram-police/internal/store"
)

const sessionCookieName = "telepolice_admin_session"

// Handler handles administrative web requests and PIN authentication.
type Handler struct {
	pin       string
	store     *store.Store
	mux       *http.ServeMux
	sessions  map[string]time.Time
	sessionMu sync.Mutex
}

// NewHandler creates an admin Handler with authenticated routes.
func NewHandler(pin string, store *store.Store) *Handler {
	h := &Handler{
		pin:      pin,
		store:    store,
		mux:      http.NewServeMux(),
		sessions: make(map[string]time.Time),
	}

	h.registerRoutes()
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) registerRoutes() {
	h.mux.HandleFunc("/admin/login", h.handleLogin)
	h.mux.HandleFunc("/admin/logout", h.handleLogout)
	h.mux.HandleFunc("/admin", h.requireAuth(h.handleDashboard))
	h.mux.HandleFunc("/admin/words", h.requireAuth(h.handleAddWords))
	h.mux.HandleFunc("/admin/words/delete", h.requireAuth(h.handleDeleteWord))
	h.mux.HandleFunc("/admin/title-words", h.requireAuth(h.handleAddTitleWords))
	h.mux.HandleFunc("/admin/title-words/delete", h.requireAuth(h.handleDeleteTitleWord))
	h.mux.HandleFunc("/admin/whitelist", h.requireAuth(h.handleAddWhitelist))
	h.mux.HandleFunc("/admin/whitelist/delete", h.requireAuth(h.handleDeleteWhitelist))
}

func (h *Handler) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !h.isAuthenticated(r) {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func (h *Handler) isAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return false
	}

	h.sessionMu.Lock()
	defer h.sessionMu.Unlock()

	exp, ok := h.sessions[cookie.Value]
	if !ok || time.Now().After(exp) {
		delete(h.sessions, cookie.Value)
		return false
	}
	return true
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if h.isAuthenticated(r) {
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		}
		_ = loginTmpl.Execute(w, nil)
		return
	}

	if r.Method == http.MethodPost {
		enteredPIN := strings.TrimSpace(r.FormValue("pin"))
		if enteredPIN == "" || enteredPIN != h.pin {
			w.WriteHeader(http.StatusUnauthorized)
			_ = loginTmpl.Execute(w, map[string]string{"Error": "PIN salah. Silakan coba lagi."})
			return
		}

		token := generateToken()
		h.sessionMu.Lock()
		h.sessions[token] = time.Now().Add(24 * time.Hour)
		h.sessionMu.Unlock()

		http.SetCookie(w, &http.Cookie{
			Name:     sessionCookieName,
			Value:    token,
			Path:     "/admin",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   86400,
		})

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil {
		h.sessionMu.Lock()
		delete(h.sessions, cookie.Value)
		h.sessionMu.Unlock()
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/admin",
		HttpOnly: true,
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

func (h *Handler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	words := h.store.GetForbiddenWords()
	titleWords := h.store.GetForbiddenTitleWords()
	whitelist, err := h.store.GetWhitelistSenders()
	if err != nil {
		slog.Error("failed to get whitelist", "error", err)
	}

	data := DashboardData{
		ForbiddenWords:      words,
		ForbiddenTitleWords: titleWords,
		WhitelistSenders:    whitelist,
		TotalWords:          len(words),
		TotalTitleWords:     len(titleWords),
		TotalWhitelist:      len(whitelist),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = dashboardTmpl.Execute(w, data)
}

func (h *Handler) handleAddWords(w http.ResponseWriter, r *http.Request) {
	raw := r.FormValue("words")
	parts := strings.Split(raw, ",")
	for _, p := range parts {
		if word := strings.TrimSpace(p); word != "" {
			_ = h.store.AddForbiddenWord(word)
		}
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handler) handleDeleteWord(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.FormValue("id"), 10, 64)
	if err == nil {
		_ = h.store.DeleteForbiddenWord(id)
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handler) handleAddTitleWords(w http.ResponseWriter, r *http.Request) {
	raw := r.FormValue("words")
	parts := strings.Split(raw, ",")
	for _, p := range parts {
		if word := strings.TrimSpace(p); word != "" {
			_ = h.store.AddForbiddenTitleWord(word)
		}
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handler) handleDeleteTitleWord(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.FormValue("id"), 10, 64)
	if err == nil {
		_ = h.store.DeleteForbiddenTitleWord(id)
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handler) handleAddWhitelist(w http.ResponseWriter, r *http.Request) {
	uid, err := strconv.ParseInt(strings.TrimSpace(r.FormValue("user_id")), 10, 64)
	if err == nil && uid > 0 {
		desc := strings.TrimSpace(r.FormValue("description"))
		_ = h.store.AddWhitelistSender(uid, desc)
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handler) handleDeleteWhitelist(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.FormValue("id"), 10, 64)
	if err == nil {
		_ = h.store.DeleteWhitelistSender(id)
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func generateToken() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
