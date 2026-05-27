package filter

import (
	"github.com/andisiahaan/telegram-police/internal/botctx"
	"github.com/andisiahaan/telegram-police/internal/config"
)

// NewChatMembersFilter removes the "X joined the group" service messages.
type NewChatMembersFilter struct{}

func (f *NewChatMembersFilter) Name() string { return "new_chat_members" }

func (f *NewChatMembersFilter) Enabled(cfg *config.Config) bool {
	return cfg.Filters.DeleteNewChatMembers
}

func (f *NewChatMembersFilter) Check(ctx *botctx.Context) (bool, string) {
	if ctx.Update.Message == nil {
		return false, ""
	}
	if len(ctx.Update.Message.NewChatMembers) > 0 {
		return true, "join notification"
	}
	return false, ""
}
