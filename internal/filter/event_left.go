package filter

import (
	"github.com/andisiahaan/telegram-police/internal/botctx"
	"github.com/andisiahaan/telegram-police/internal/config"
)

// LeftChatMemberFilter removes the "X left the group" service messages.
type LeftChatMemberFilter struct{}

func (f *LeftChatMemberFilter) Name() string { return "left_chat_member" }

func (f *LeftChatMemberFilter) Enabled(cfg *config.Config) bool {
	return cfg.Filters.DeleteLeftChatMember
}

func (f *LeftChatMemberFilter) Check(ctx *botctx.Context) (bool, string) {
	if ctx.Update.Message == nil {
		return false, ""
	}
	if ctx.Update.Message.LeftChatMember != nil {
		return true, "leave notification"
	}
	return false, ""
}
