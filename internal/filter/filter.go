package filter

import (
	"github.com/andisiahaan/telegram-police/internal/botctx"
	"github.com/andisiahaan/telegram-police/internal/config"
)

// Filter is the interface every filter must implement.
type Filter interface {
	// Name returns a short identifier used in logs and violation reasons.
	Name() string
	// Enabled reports whether this filter should run given the current config.
	Enabled(cfg *config.Config) bool
	// Check inspects the message and returns whether it violated this filter.
	Check(ctx *botctx.Context) (violated bool, reason string)
}

// RunChain runs all enabled filters against ctx. It does not stop on the first
// violation — all reasons are accumulated in ctx.Violations.
func RunChain(filters []Filter, ctx *botctx.Context) {
	for _, f := range filters {
		if !f.Enabled(ctx.Config) {
			continue
		}
		if violated, reason := f.Check(ctx); violated {
			ctx.AddViolation(reason)
		}
	}
}
