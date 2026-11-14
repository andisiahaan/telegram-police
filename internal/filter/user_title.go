package filter

import (
	"fmt"
	"strings"

	"github.com/andisiahaan/telegram-police/internal/botctx"
	"github.com/andisiahaan/telegram-police/internal/config"
	"github.com/andisiahaan/telegram-police/internal/store"
)

// UserTitleFilter rejects messages from users whose display name or username
// contains a word on the forbidden title list.
type UserTitleFilter struct {
	store *store.Store
}

// NewUserTitleFilter creates a dynamic title words filter backed by the store.
func NewUserTitleFilter(s *store.Store) *UserTitleFilter {
	return &UserTitleFilter{store: s}
}

func (f *UserTitleFilter) Name() string { return "user_title" }

func (f *UserTitleFilter) Enabled(cfg *config.Config) bool {
	if f.store != nil {
		return len(f.store.GetForbiddenTitleWords()) > 0
	}
	return len(cfg.ForbiddenTitleWords) > 0
}

func (f *UserTitleFilter) Check(ctx *botctx.Context) (bool, string) {
	if ctx.Update.Message == nil || ctx.Update.Message.From == nil {
		return false, ""
	}

	user := ctx.Update.Message.From
	combined := strings.ToLower(strings.Join([]string{
		user.FirstName,
		user.LastName,
		user.Username,
	}, " "))

	if f.store != nil {
		for _, item := range f.store.GetForbiddenTitleWords() {
			if strings.Contains(combined, strings.ToLower(item.Word)) {
				return true, fmt.Sprintf("display name contains forbidden word: %q", item.Word)
			}
		}
		return false, ""
	}

	for _, word := range ctx.Config.ForbiddenTitleWords {
		if strings.Contains(combined, strings.ToLower(word)) {
			return true, fmt.Sprintf("display name contains forbidden word: %q", word)
		}
	}
	return false, ""
}
