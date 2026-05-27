package ratelimit

import (
	"time"

	"github.com/andisiahaan/telegram-police/internal/botctx"
	"github.com/andisiahaan/telegram-police/internal/config"
)

// FloodChecker detects users sending messages too fast.
type FloodChecker struct {
	window *SlidingWindow
}

// NewFloodChecker builds a FloodChecker from the flood section of the config.
func NewFloodChecker(cfg *config.Config) *FloodChecker {
	interval := time.Duration(cfg.Flood.IntervalMinutes) * time.Minute
	return &FloodChecker{
		window: NewSlidingWindow(cfg.Flood.MaxMessages, interval),
	}
}

// Check records the message and returns true if the user has exceeded the flood limit.
func (fc *FloodChecker) Check(ctx *botctx.Context) bool {
	fc.window.Add(ctx.ChatID, ctx.UserID)
	return fc.window.Exceeded(ctx.ChatID, ctx.UserID)
}
