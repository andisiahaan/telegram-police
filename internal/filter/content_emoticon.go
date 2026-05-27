package filter

import (
	"fmt"
	"unicode"

	"github.com/andisiahaan/telegram-police/internal/botctx"
	"github.com/andisiahaan/telegram-police/internal/config"
)

// EmoticonFilter rejects messages with too many emoji.
type EmoticonFilter struct{}

func (f *EmoticonFilter) Name() string { return "emoticon" }

func (f *EmoticonFilter) Enabled(cfg *config.Config) bool {
	return cfg.MaxEmoticons > 0
}

func (f *EmoticonFilter) Check(ctx *botctx.Context) (bool, string) {
	if ctx.Update.Message == nil {
		return false, ""
	}
	text := ctx.Update.Message.Text
	if text == "" {
		return false, ""
	}

	count := countEmojis(text)
	if count > ctx.Config.MaxEmoticons {
		return true, fmt.Sprintf("too many emoji (%d/%d)", count, ctx.Config.MaxEmoticons)
	}
	return false, ""
}

func countEmojis(text string) int {
	count := 0
	for _, r := range text {
		if isEmoji(r) {
			count++
		}
	}
	return count
}

func isEmoji(r rune) bool {
	return unicode.Is(unicode.So, r) ||
		(r >= 0x1F300 && r <= 0x1FAFF) ||
		(r >= 0x1F900 && r <= 0x1F9FF) ||
		(r >= 0x2600 && r <= 0x27BF) ||
		(r >= 0xFE00 && r <= 0xFE0F)
}
