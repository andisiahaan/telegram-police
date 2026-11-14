package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreCRUD(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	initWords := []string{"badword", "spam"}
	initTitleWords := []string{"admin_fake"}
	initWhitelist := []int64{111222, 333444}

	s, err := New(dbPath, initWords, initTitleWords, initWhitelist)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	// 1. Verify initial seed
	words := s.GetForbiddenWords()
	if len(words) != 2 {
		t.Errorf("expected 2 words, got %d", len(words))
	}
	if s.GetForbiddenRegex() == nil {
		t.Errorf("expected compiled regex, got nil")
	}

	// 2. Add word
	if err := s.AddForbiddenWord("crypto"); err != nil {
		t.Fatalf("failed to add word: %v", err)
	}
	words = s.GetForbiddenWords()
	if len(words) != 3 {
		t.Errorf("expected 3 words, got %d", len(words))
	}

	// 3. Delete word
	var cryptoID int64
	for _, w := range words {
		if w.Word == "crypto" {
			cryptoID = w.ID
			break
		}
	}
	if err := s.DeleteForbiddenWord(cryptoID); err != nil {
		t.Fatalf("failed to delete word: %v", err)
	}
	words = s.GetForbiddenWords()
	if len(words) != 2 {
		t.Errorf("expected 2 words after delete, got %d", len(words))
	}

	// 4. Whitelist check
	if !s.IsWhitelisted(111222) {
		t.Errorf("expected 111222 to be whitelisted")
	}
	if s.IsWhitelisted(999999) {
		t.Errorf("expected 999999 NOT to be whitelisted")
	}

	// 5. Add whitelist
	if err := s.AddWhitelistSender(999999, "test user"); err != nil {
		t.Fatalf("failed to add whitelist: %v", err)
	}
	if !s.IsWhitelisted(999999) {
		t.Errorf("expected 999999 to be whitelisted after add")
	}

	_ = os.Remove(dbPath)
}
