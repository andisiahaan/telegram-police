package ratelimit

import (
	"time"

	"github.com/andisiahaan/telegram-police/internal/botctx"
	"github.com/andisiahaan/telegram-police/internal/config"
)

// SpamChecker tracks accumulated filter violations per user per chat.
type SpamChecker struct {
	window *SlidingWindow
}

// NewSpamChecker builds a SpamChecker from the spam section of the config.
func NewSpamChecker(cfg *config.Config) *SpamChecker {
	interval := time.Duration(cfg.Spam.WindowMinutes) * time.Minute
	return &SpamChecker{
		window: NewSlidingWindow(cfg.Spam.MaxViolations, interval),
	}
}

// AddViolation records one filter violation for this user.
func (sc *SpamChecker) AddViolation(ctx *botctx.Context) {
	sc.window.Add(ctx.ChatID, ctx.UserID)
}

// Check returns true if the user's violation count has exceeded the spam threshold.
func (sc *SpamChecker) Check(ctx *botctx.Context) bool {
	return sc.window.Exceeded(ctx.ChatID, ctx.UserID)
}
