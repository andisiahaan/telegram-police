package ratelimit

import (
	"fmt"
	"sync"
	"time"
)

// SlidingWindow is a thread-safe per-key event counter using a sliding time window.
// Used for both flood detection and spam violation counting.
type SlidingWindow struct {
	mu       sync.Mutex
	entries  map[string][]time.Time
	maxCount int
	interval time.Duration
}

// NewSlidingWindow creates a SlidingWindow with the given limit and window size.
func NewSlidingWindow(maxCount int, interval time.Duration) *SlidingWindow {
	return &SlidingWindow{
		entries:  make(map[string][]time.Time),
		maxCount: maxCount,
		interval: interval,
	}
}

// Add records a new event for the given chat/user pair and prunes expired ones.
func (sw *SlidingWindow) Add(chatID, userID int64) {
	key := buildKey(chatID, userID)
	now := time.Now()

	sw.mu.Lock()
	defer sw.mu.Unlock()

	sw.entries[key] = append(prune(sw.entries[key], now, sw.interval), now)
}

// Count returns the number of events still within the window for a given key.
func (sw *SlidingWindow) Count(chatID, userID int64) int {
	key := buildKey(chatID, userID)
	now := time.Now()

	sw.mu.Lock()
	defer sw.mu.Unlock()

	sw.entries[key] = prune(sw.entries[key], now, sw.interval)
	return len(sw.entries[key])
}

// Exceeded reports whether the event count has reached or passed maxCount.
func (sw *SlidingWindow) Exceeded(chatID, userID int64) bool {
	return sw.Count(chatID, userID) >= sw.maxCount
}

// prune removes timestamps that have fallen outside the window.
func prune(times []time.Time, now time.Time, interval time.Duration) []time.Time {
	cutoff := now.Add(-interval)
	i := 0
	for i < len(times) && times[i].Before(cutoff) {
		i++
	}
	return times[i:]
}

func buildKey(chatID, userID int64) string {
	return fmt.Sprintf("%d:%d", chatID, userID)
}
