package filter

import (
	"fmt"
	"strings"

	"github.com/andisiahaan/telegram-police/internal/botctx"
	"github.com/andisiahaan/telegram-police/internal/config"
)

// UserTitleFilter rejects messages from users whose display name or username
// contains a word on the forbidden title list.
type UserTitleFilter struct{}

func (f *UserTitleFilter) Name() string { return "user_title" }

func (f *UserTitleFilter) Enabled(cfg *config.Config) bool {
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

	for _, word := range ctx.Config.ForbiddenTitleWords {
		if strings.Contains(combined, strings.ToLower(word)) {
			return true, fmt.Sprintf("display name contains forbidden word: %q", word)
		}
	}
	return false, ""
}
