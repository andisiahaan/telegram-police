package filter

import (
	"fmt"
	"strings"

	"github.com/andisiahaan/telegram-police/internal/botctx"
	"github.com/andisiahaan/telegram-police/internal/config"
)

// ForbiddenWordsFilter rejects messages that contain any word on the blocklist.
type ForbiddenWordsFilter struct{}

func (f *ForbiddenWordsFilter) Name() string { return "forbidden_words" }

func (f *ForbiddenWordsFilter) Enabled(cfg *config.Config) bool {
	return len(cfg.ForbiddenWords) > 0
}

func (f *ForbiddenWordsFilter) Check(ctx *botctx.Context) (bool, string) {
	if ctx.Update.Message == nil {
		return false, ""
	}
	text := strings.ToLower(ctx.Update.Message.Text)
	if text == "" {
		return false, ""
	}

	for _, word := range ctx.Config.ForbiddenWords {
		if strings.Contains(text, strings.ToLower(word)) {
			return true, fmt.Sprintf("contains forbidden word: %q", word)
		}
	}
	return false, ""
}
