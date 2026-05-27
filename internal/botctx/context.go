// Package botctx defines the Context struct that flows through the entire bot pipeline.
// It intentionally does not import other internal packages to avoid circular imports.
package botctx

import (
	"github.com/andisiahaan/telegram-police/internal/config"
	"github.com/andisiahaan/telegram-police/internal/telegram"
)

// Context carries all the data a pipeline stage needs to do its job.
type Context struct {
	Update     *telegram.Update
	Config     *config.Config
	ChatID     int64
	UserID     int64
	Username   string
	MessageID       int
	Violations      []string // collected by the filter chain; empty means no violations
	ExecutedActions []string // actions that were successfully executed
	Bypassed        bool
	BypassReason    string
}

// New builds a Context from a Telegram Update, extracting the most-used fields
// as shortcuts so every stage doesn't have to nil-check the whole chain.
func New(update *telegram.Update, cfg *config.Config) *Context {
	ctx := &Context{
		Update:     update,
		Config:          cfg,
		Violations:      []string{},
		ExecutedActions: []string{},
	}

	if update.Message != nil {
		ctx.MessageID = update.Message.MessageID
		if update.Message.Chat != nil {
			ctx.ChatID = update.Message.Chat.ID
		}
		if update.Message.From != nil {
			ctx.UserID = update.Message.From.ID
			ctx.Username = update.Message.From.Username
			if ctx.Username == "" {
				ctx.Username = update.Message.From.FirstName
			}
		}
	}

	return ctx
}

// AddViolation appends a reason string to the violation list.
func (c *Context) AddViolation(reason string) {
	c.Violations = append(c.Violations, reason)
}

// HasViolations reports whether any violations have been recorded.
func (c *Context) HasViolations() bool {
	return len(c.Violations) > 0
}

// AddAction appends an action name to the executed actions list.
func (c *Context) AddAction(action string) {
	c.ExecutedActions = append(c.ExecutedActions, action)
}
