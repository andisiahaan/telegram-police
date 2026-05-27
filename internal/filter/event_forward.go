package filter

import (
	"github.com/andisiahaan/telegram-police/internal/botctx"
	"github.com/andisiahaan/telegram-police/internal/config"
)

// ForwardedMessageFilter deletes messages that were forwarded from another user or channel.
type ForwardedMessageFilter struct{}

func (f *ForwardedMessageFilter) Name() string { return "forwarded_message" }

func (f *ForwardedMessageFilter) Enabled(cfg *config.Config) bool {
	return cfg.Filters.DeleteForwardedMessages
}

func (f *ForwardedMessageFilter) Check(ctx *botctx.Context) (bool, string) {
	if ctx.Update.Message == nil {
		return false, ""
	}
	msg := ctx.Update.Message
	if msg.ForwardFrom != nil || msg.ForwardFromChat != nil {
		return true, "forwarded message"
	}
	return false, ""
}
