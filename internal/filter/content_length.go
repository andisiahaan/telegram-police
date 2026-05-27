package filter

import (
	"fmt"

	"github.com/andisiahaan/telegram-police/internal/botctx"
	"github.com/andisiahaan/telegram-police/internal/config"
)

// ContentLengthFilter rejects messages that exceed the character limit.
type ContentLengthFilter struct{}

func (f *ContentLengthFilter) Name() string { return "content_length" }

func (f *ContentLengthFilter) Enabled(cfg *config.Config) bool {
	return cfg.MaxCharacters > 0
}

func (f *ContentLengthFilter) Check(ctx *botctx.Context) (bool, string) {
	if ctx.Update.Message == nil {
		return false, ""
	}
	text := ctx.Update.Message.Text
	if text == "" {
		return false, ""
	}
	// Count runes, not bytes, so multi-byte characters (emoji, CJK, etc.) are counted correctly.
	length := len([]rune(text))
	if length > ctx.Config.MaxCharacters {
		return true, fmt.Sprintf("message too long (%d/%d chars)", length, ctx.Config.MaxCharacters)
	}
	return false, ""
}
