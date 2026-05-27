package cache

import (
	"fmt"
	"time"
)

// IsAdmin checks whether a user is an admin or creator in a chat, using the cache.
// Returns (isAdmin, found). A false found means cache miss — caller should hit the API.
func IsAdmin(store *Store, chatID, userID int64) (isAdmin bool, found bool) {
	key := adminKey(chatID, userID)
	val, ok := store.Get(key)
	if !ok {
		return false, false
	}
	isAdmin, _ = val.(bool)
	return isAdmin, true
}

// SetAdmin stores a user's admin status in the cache with the given TTL.
func SetAdmin(store *Store, chatID, userID int64, isAdmin bool, ttl time.Duration) {
	store.Set(adminKey(chatID, userID), isAdmin, ttl)
}

func adminKey(chatID, userID int64) string {
	return fmt.Sprintf("admin:%d:%d", chatID, userID)
}
