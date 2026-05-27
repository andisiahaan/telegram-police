package action

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/andisiahaan/telegram-police/internal/botctx"
	"github.com/andisiahaan/telegram-police/internal/telegram"
)

// WarnAction sends a warning message to the chat, optionally auto-deleting it after a delay.
type WarnAction struct {
	api *telegram.API
}

// NewWarnAction creates a WarnAction.
func NewWarnAction(api *telegram.API) *WarnAction {
	return &WarnAction{api: api}
}

func (a *WarnAction) Name() string { return "warn" }

func (a *WarnAction) Execute(ctx *botctx.Context) error {
	cfg := ctx.Config
	if !cfg.WarnEnabled {
		return nil
	}

	msg := buildWarnMessage(cfg.WarnMessageTemplate, ctx.Username, ctx.Violations)

	msgID, err := a.api.SendMessage(ctx.ChatID, msg)
	if err != nil {
		slog.Error("failed to send warning message",
			"chat_id", ctx.ChatID,
			"user_id", ctx.UserID,
			"error", err,
		)
		return fmt.Errorf("warn action: %w", err)
	}

	slog.Info("warning sent",
		"chat_id", ctx.ChatID,
		"user_id", ctx.UserID,
		"warn_message_id", msgID,
	)

	if cfg.WarnAutoDeleteSeconds > 0 {
		go func(chatID int64, warnMsgID int, delay int) {
			time.Sleep(time.Duration(delay) * time.Second)
			if delErr := a.api.DeleteMessage(chatID, warnMsgID); delErr != nil {
				slog.Error("failed to auto-delete warning message",
					"chat_id", chatID,
					"warn_message_id", warnMsgID,
					"error", delErr,
				)
			}
		}(ctx.ChatID, msgID, cfg.WarnAutoDeleteSeconds)
	}

	return nil
}

// buildWarnMessage replaces {username} and {reason} placeholders in the template.
func buildWarnMessage(template, username string, violations []string) string {
	reason := strings.Join(violations, ", ")
	if reason == "" {
		reason = "violated group rules"
	}
	msg := strings.ReplaceAll(template, "{username}", username)
	msg = strings.ReplaceAll(msg, "{reason}", reason)
	return msg
}
