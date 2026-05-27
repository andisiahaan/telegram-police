package filter

import (
	"fmt"
	"strings"

	"github.com/andisiahaan/telegram-police/internal/botctx"
	"github.com/andisiahaan/telegram-police/internal/config"
	"github.com/andisiahaan/telegram-police/internal/telegram"
)

// MentionFilter rejects messages with too many @mentions.
type MentionFilter struct{}

func (f *MentionFilter) Name() string { return "mention" }

func (f *MentionFilter) Enabled(cfg *config.Config) bool {
	return cfg.MaxMentions > 0
}

func (f *MentionFilter) Check(ctx *botctx.Context) (bool, string) {
	if ctx.Update.Message == nil {
		return false, ""
	}

	count := countMentions(ctx.Update.Message)
	if count > ctx.Config.MaxMentions {
		return true, fmt.Sprintf("too many mentions (%d/%d)", count, ctx.Config.MaxMentions)
	}
	return false, ""
}

// countMentions reads from Telegram's Entities array when available (accurate),
// and falls back to counting "@" occurrences in the raw text.
func countMentions(msg *telegram.Message) int {
	if len(msg.Entities) > 0 {
		count := 0
		for _, e := range msg.Entities {
			if e.Type == "mention" || e.Type == "text_mention" {
				count++
			}
		}
		return count
	}

	if msg.Text == "" {
		return 0
	}
	return strings.Count(msg.Text, "@")
}
