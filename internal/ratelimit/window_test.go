package ratelimit

import (
	"testing"
	"time"
)

func TestSlidingWindow(t *testing.T) {
	sw := NewSlidingWindow(3, 100*time.Millisecond)

	// User sends 3 messages quickly
	sw.Add(1, 1)
	sw.Add(1, 1)
	sw.Add(1, 1)

	if !sw.Exceeded(1, 1) {
		t.Errorf("Expected window to be exceeded after 3 messages")
	}

	// Wait for window to slide
	time.Sleep(150 * time.Millisecond)

	if sw.Exceeded(1, 1) {
		t.Errorf("Expected window to not be exceeded after waiting")
	}
	
	if sw.Count(1, 1) != 0 {
		t.Errorf("Expected count to be 0 after waiting, got %d", sw.Count(1, 1))
	}
}
