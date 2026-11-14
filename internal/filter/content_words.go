package filter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/andisiahaan/telegram-police/internal/botctx"
	"github.com/andisiahaan/telegram-police/internal/config"
	"github.com/andisiahaan/telegram-police/internal/store"
)

// ForbiddenWordsFilter rejects messages that contain any word on the blocklist.
type ForbiddenWordsFilter struct {
	store       *store.Store
	staticRegex *regexp.Regexp
}

// NewForbiddenWordsFilter creates a filter backed by the dynamic WordStore.
func NewForbiddenWordsFilter(s *store.Store) *ForbiddenWordsFilter {
	return &ForbiddenWordsFilter{store: s}
}

// NewStaticForbiddenWordsFilter creates a filter with static words (used for testing).
func NewStaticForbiddenWordsFilter(words []string) *ForbiddenWordsFilter {
	if len(words) == 0 {
		return &ForbiddenWordsFilter{}
	}
	var escaped []string
	for _, w := range words {
		escaped = append(escaped, regexp.QuoteMeta(strings.ToLower(w)))
	}
	pattern := fmt.Sprintf(`(?i)(?:^|\W)(%s)(?:\W|$)`, strings.Join(escaped, "|"))
	return &ForbiddenWordsFilter{
		staticRegex: regexp.MustCompile(pattern),
	}
}

func (f *ForbiddenWordsFilter) Name() string { return "forbidden_words" }

func (f *ForbiddenWordsFilter) getRegex() *regexp.Regexp {
	if f.store != nil {
		return f.store.GetForbiddenRegex()
	}
	return f.staticRegex
}

func (f *ForbiddenWordsFilter) Enabled(cfg *config.Config) bool {
	return f.getRegex() != nil
}

func (f *ForbiddenWordsFilter) Check(ctx *botctx.Context) (bool, string) {
	if ctx.Update.Message == nil {
		return false, ""
	}
	regex := f.getRegex()
	if regex == nil {
		return false, ""
	}
	text := ctx.Update.Message.Text
	if text == "" {
		return false, ""
	}

	match := regex.FindStringSubmatch(text)
	if match != nil && len(match) > 1 {
		return true, fmt.Sprintf("contains forbidden word: %q", strings.ToLower(match[1]))
	}
	return false, ""
}
