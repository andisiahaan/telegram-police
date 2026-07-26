package filter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/andisiahaan/telegram-police/internal/botctx"
	"github.com/andisiahaan/telegram-police/internal/config"
)

// ForbiddenWordsFilter rejects messages that contain any word on the blocklist.
type ForbiddenWordsFilter struct {
	regex *regexp.Regexp
}

// NewForbiddenWordsFilter precompiles the forbidden words into a single regex.
func NewForbiddenWordsFilter(words []string) *ForbiddenWordsFilter {
	if len(words) == 0 {
		return &ForbiddenWordsFilter{}
	}
	var escaped []string
	for _, w := range words {
		escaped = append(escaped, regexp.QuoteMeta(strings.ToLower(w)))
	}
	pattern := fmt.Sprintf(`(?i)(?:^|\W)(%s)(?:\W|$)`, strings.Join(escaped, "|"))
	return &ForbiddenWordsFilter{
		regex: regexp.MustCompile(pattern),
	}
}

func (f *ForbiddenWordsFilter) Name() string { return "forbidden_words" }

func (f *ForbiddenWordsFilter) Enabled(cfg *config.Config) bool {
	return f.regex != nil && len(cfg.ForbiddenWords) > 0
}

func (f *ForbiddenWordsFilter) Check(ctx *botctx.Context) (bool, string) {
	if ctx.Update.Message == nil || f.regex == nil {
		return false, ""
	}
	text := ctx.Update.Message.Text
	if text == "" {
		return false, ""
	}

	match := f.regex.FindStringSubmatch(text)
	if match != nil && len(match) > 1 {
		return true, fmt.Sprintf("contains forbidden word: %q", strings.ToLower(match[1]))
	}
	return false, ""
}
