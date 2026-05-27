package filter

import (
	"fmt"
	"unicode"

	"github.com/andisiahaan/telegram-police/internal/botctx"
	"github.com/andisiahaan/telegram-police/internal/config"
)

// NonLatinFilter rejects messages where non-latin letters exceed a configured percentage.
// Checked scripts: Arabic, Cyrillic, Han (CJK), Devanagari, Thai, Hebrew, Hangul, Hiragana, Katakana.
type NonLatinFilter struct{}

func (f *NonLatinFilter) Name() string { return "non_latin" }

func (f *NonLatinFilter) Enabled(cfg *config.Config) bool {
	return cfg.Filters.CheckNonLatin
}

func (f *NonLatinFilter) Check(ctx *botctx.Context) (bool, string) {
	if ctx.Update.Message == nil {
		return false, ""
	}
	text := ctx.Update.Message.Text
	if text == "" {
		return false, ""
	}

	totalLetters := 0
	nonLatinCount := 0

	for _, r := range []rune(text) {
		if !unicode.IsLetter(r) {
			continue
		}
		totalLetters++
		if isNonLatin(r) {
			nonLatinCount++
		}
	}

	if totalLetters == 0 {
		return false, ""
	}

	threshold := ctx.Config.NonLatinThresholdPercent
	percent := nonLatinCount * 100 / totalLetters
	if percent > threshold {
		return true, fmt.Sprintf("non-latin ratio too high (%d%%/%d%%)", percent, threshold)
	}
	return false, ""
}

func isNonLatin(r rune) bool {
	return unicode.In(r,
		unicode.Arabic,
		unicode.Cyrillic,
		unicode.Han,
		unicode.Devanagari,
		unicode.Thai,
		unicode.Hebrew,
		unicode.Hangul,
		unicode.Hiragana,
		unicode.Katakana,
		unicode.Georgian,
		unicode.Armenian,
	)
}
