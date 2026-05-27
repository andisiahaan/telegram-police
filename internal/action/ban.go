package action

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/andisiahaan/telegram-police/internal/botctx"
	"github.com/andisiahaan/telegram-police/internal/telegram"
)

// BanAction bans the sender from the chat for the configured duration.
type BanAction struct {
	api *telegram.API
}

// NewBanAction creates a BanAction.
func NewBanAction(api *telegram.API) *BanAction {
	return &BanAction{api: api}
}

func (a *BanAction) Name() string { return "ban" }

func (a *BanAction) Execute(ctx *botctx.Context) error {
	days := ctx.Config.BanDurationDays
	var untilDate int64
	if days > 0 {
		untilDate = time.Now().Add(time.Duration(days) * 24 * time.Hour).Unix()
	}

	err := a.api.BanChatMember(ctx.ChatID, ctx.UserID, untilDate)
	if err != nil {
		slog.Error("failed to ban user",
			"chat_id", ctx.ChatID,
			"user_id", ctx.UserID,
			"error", err,
		)
		return fmt.Errorf("ban action: %w", err)
	}
	slog.Info("user banned",
		"chat_id", ctx.ChatID,
		"user_id", ctx.UserID,
		"duration_days", days,
	)
	return nil
}
