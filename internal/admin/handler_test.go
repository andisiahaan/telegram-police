package admin

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andisiahaan/telegram-police/internal/store"
)

func setupTestStore(t *testing.T) *store.Store {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "admin_test.db")
	s, err := store.New(dbPath, []string{"foo", "bar"}, nil, nil)
	if err != nil {
		t.Fatalf("failed to create test store: %v", err)
	}
	return s
}

func TestAdminAuthAndCRUD(t *testing.T) {
	s := setupTestStore(t)
	defer s.Close()

	pin := "987654"
	handler := NewHandler(pin, s)

	// 1. Unauthenticated access to /admin should redirect to /admin/login
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/admin/login" {
		t.Errorf("expected redirect to /admin/login, got status %d location %s", w.Code, w.Header().Get("Location"))
	}

	// 2. Login with wrong PIN should return 401
	form := url.Values{"pin": {"wrongpin"}}
	req = httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong pin, got %d", w.Code)
	}

	// 3. Login with correct PIN should succeed and return session cookie
	form = url.Values{"pin": {pin}}
	req = httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/admin" {
		t.Errorf("expected redirect to /admin, got status %d", w.Code)
	}

	cookie := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookie {
		if c.Name == sessionCookieName {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil || sessionCookie.Value == "" {
		t.Fatalf("session cookie not found after successful login")
	}

	// 4. Authenticated request to /admin should return 200 OK
	req = httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.AddCookie(sessionCookie)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for authenticated dashboard, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "TelePolice Admin") {
		t.Errorf("dashboard HTML missing title")
	}

	// 5. Add words via /admin/words
	form = url.Values{"words": {"crypto, bitcoin"}}
	req = httptest.NewRequest(http.MethodPost, "/admin/words", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusSeeOther {
		t.Errorf("expected redirect after adding words, got %d", w.Code)
	}

	words := s.GetForbiddenWords()
	foundCrypto := false
	for _, item := range words {
		if item.Word == "crypto" {
			foundCrypto = true
			break
		}
	}
	if !foundCrypto {
		t.Errorf("expected 'crypto' to be added to store")
	}
}
