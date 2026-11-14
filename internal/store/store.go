package store

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	_ "modernc.org/sqlite"
)

// Store manages SQLite persistence for keywords and whitelisted senders
// while maintaining a fast, thread-safe in-memory cache for message filtering.
type Store struct {
	db *sql.DB
	mu sync.RWMutex

	wordsCached      []WordItem
	titleWordsCached []TitleWordItem
	whitelistCached  map[int64]string
	regexCached      *regexp.Regexp
}

// New opens the SQLite database, runs migrations, seeds initial data if empty,
// and loads the in-memory cache.
func New(dbPath string, initWords, initTitleWords []string, initWhitelist []int64) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("store: create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("store: open sqlite: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("store: ping sqlite: %w", err)
	}

	s := &Store{
		db:              db,
		whitelistCached: make(map[int64]string),
	}

	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("store: migration: %w", err)
	}

	if err := s.seed(initWords, initTitleWords, initWhitelist); err != nil {
		slog.Warn("store: seed warning", "error", err)
	}

	if err := s.Reload(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("store: reload cache: %w", err)
	}

	return s, nil
}

func (s *Store) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS forbidden_words (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		word TEXT UNIQUE NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS forbidden_title_words (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		word TEXT UNIQUE NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS whitelist_senders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER UNIQUE NOT NULL,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *Store) seed(initWords, initTitleWords []string, initWhitelist []int64) error {
	var count int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM forbidden_words").Scan(&count); err == nil && count == 0 {
		for _, w := range initWords {
			clean := strings.TrimSpace(strings.ToLower(w))
			if clean != "" {
				_, _ = s.db.Exec("INSERT OR IGNORE INTO forbidden_words (word) VALUES (?)", clean)
			}
		}
	}

	if err := s.db.QueryRow("SELECT COUNT(*) FROM forbidden_title_words").Scan(&count); err == nil && count == 0 {
		for _, w := range initTitleWords {
			clean := strings.TrimSpace(strings.ToLower(w))
			if clean != "" {
				_, _ = s.db.Exec("INSERT OR IGNORE INTO forbidden_title_words (word) VALUES (?)", clean)
			}
		}
	}

	if err := s.db.QueryRow("SELECT COUNT(*) FROM whitelist_senders").Scan(&count); err == nil && count == 0 {
		for _, id := range initWhitelist {
			if id > 0 {
				_, _ = s.db.Exec("INSERT OR IGNORE INTO whitelist_senders (user_id, description) VALUES (?, ?)", id, "Initial config seed")
			}
		}
	}

	return nil
}

// Reload refreshes the in-memory cache and re-compiles the regex.
func (s *Store) Reload() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Forbidden words
	wordRows, err := s.db.Query("SELECT id, word, created_at FROM forbidden_words ORDER BY word ASC")
	if err != nil {
		return err
	}
	defer wordRows.Close()

	var words []WordItem
	var wordStrings []string
	for wordRows.Next() {
		var item WordItem
		if err := wordRows.Scan(&item.ID, &item.Word, &item.CreatedAt); err != nil {
			return err
		}
		words = append(words, item)
		wordStrings = append(wordStrings, item.Word)
	}

	// 2. Forbidden title words
	titleRows, err := s.db.Query("SELECT id, word, created_at FROM forbidden_title_words ORDER BY word ASC")
	if err != nil {
		return err
	}
	defer titleRows.Close()

	var titleWords []TitleWordItem
	for titleRows.Next() {
		var item TitleWordItem
		if err := titleRows.Scan(&item.ID, &item.Word, &item.CreatedAt); err != nil {
			return err
		}
		titleWords = append(titleWords, item)
	}

	// 3. Whitelist senders
	wlRows, err := s.db.Query("SELECT id, user_id, description, created_at FROM whitelist_senders ORDER BY id ASC")
	if err != nil {
		return err
	}
	defer wlRows.Close()

	whitelistMap := make(map[int64]string)
	for wlRows.Next() {
		var item WhitelistItem
		var desc sql.NullString
		if err := wlRows.Scan(&item.ID, &item.UserID, &desc, &item.CreatedAt); err != nil {
			return err
		}
		whitelistMap[item.UserID] = desc.String
	}

	// 4. Precompile regex
	var compiled *regexp.Regexp
	if len(wordStrings) > 0 {
		var escaped []string
		for _, w := range wordStrings {
			escaped = append(escaped, regexp.QuoteMeta(strings.ToLower(w)))
		}
		pattern := fmt.Sprintf(`(?i)(?:^|\W)(%s)(?:\W|$)`, strings.Join(escaped, "|"))
		compiled = regexp.MustCompile(pattern)
	}

	s.wordsCached = words
	s.titleWordsCached = titleWords
	s.whitelistCached = whitelistMap
	s.regexCached = compiled
	return nil
}

// GetForbiddenWords returns a snapshot of all active forbidden words.
func (s *Store) GetForbiddenWords() []WordItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]WordItem, len(s.wordsCached))
	copy(res, s.wordsCached)
	return res
}

// GetForbiddenRegex returns the precompiled forbidden words regex for message matching.
func (s *Store) GetForbiddenRegex() *regexp.Regexp {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.regexCached
}

// AddForbiddenWord inserts a new forbidden word and hot-reloads the cache.
func (s *Store) AddForbiddenWord(word string) error {
	clean := strings.TrimSpace(strings.ToLower(word))
	if clean == "" {
		return fmt.Errorf("word cannot be empty")
	}
	_, err := s.db.Exec("INSERT INTO forbidden_words (word) VALUES (?)", clean)
	if err != nil {
		return err
	}
	return s.Reload()
}

// DeleteForbiddenWord deletes a forbidden word by ID and hot-reloads the cache.
func (s *Store) DeleteForbiddenWord(id int64) error {
	_, err := s.db.Exec("DELETE FROM forbidden_words WHERE id = ?", id)
	if err != nil {
		return err
	}
	return s.Reload()
}

// GetForbiddenTitleWords returns a snapshot of all active forbidden title words.
func (s *Store) GetForbiddenTitleWords() []TitleWordItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]TitleWordItem, len(s.titleWordsCached))
	copy(res, s.titleWordsCached)
	return res
}

// AddForbiddenTitleWord inserts a new forbidden title word and hot-reloads the cache.
func (s *Store) AddForbiddenTitleWord(word string) error {
	clean := strings.TrimSpace(strings.ToLower(word))
	if clean == "" {
		return fmt.Errorf("title word cannot be empty")
	}
	_, err := s.db.Exec("INSERT INTO forbidden_title_words (word) VALUES (?)", clean)
	if err != nil {
		return err
	}
	return s.Reload()
}

// DeleteForbiddenTitleWord deletes a forbidden title word by ID and hot-reloads the cache.
func (s *Store) DeleteForbiddenTitleWord(id int64) error {
	_, err := s.db.Exec("DELETE FROM forbidden_title_words WHERE id = ?", id)
	if err != nil {
		return err
	}
	return s.Reload()
}

// GetWhitelistSenders returns all whitelisted senders.
func (s *Store) GetWhitelistSenders() ([]WhitelistItem, error) {
	rows, err := s.db.Query("SELECT id, user_id, COALESCE(description, ''), created_at FROM whitelist_senders ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []WhitelistItem
	for rows.Next() {
		var item WhitelistItem
		if err := rows.Scan(&item.ID, &item.UserID, &item.Description, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// IsWhitelisted reports whether the given user ID is in the whitelist cache.
func (s *Store) IsWhitelisted(userID int64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, found := s.whitelistCached[userID]
	return found
}

// AddWhitelistSender inserts a user ID into the whitelist and hot-reloads the cache.
func (s *Store) AddWhitelistSender(userID int64, description string) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user id")
	}
	_, err := s.db.Exec("INSERT INTO whitelist_senders (user_id, description) VALUES (?, ?)", userID, description)
	if err != nil {
		return err
	}
	return s.Reload()
}

// DeleteWhitelistSender deletes a whitelist item by ID and hot-reloads the cache.
func (s *Store) DeleteWhitelistSender(id int64) error {
	_, err := s.db.Exec("DELETE FROM whitelist_senders WHERE id = ?", id)
	if err != nil {
		return err
	}
	return s.Reload()
}

// Close closes the underlying SQLite database.
func (s *Store) Close() error {
	return s.db.Close()
}
