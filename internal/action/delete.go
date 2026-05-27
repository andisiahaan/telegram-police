package action

import (
	"fmt"
	"log/slog"

	"github.com/andisiahaan/telegram-police/internal/botctx"
	"github.com/andisiahaan/telegram-police/internal/telegram"
)

// DeleteAction deletes the offending message.
type DeleteAction struct {
	api *telegram.API
}

// NewDeleteAction creates a DeleteAction.
func NewDeleteAction(api *telegram.API) *DeleteAction {
	return &DeleteAction{api: api}
}

func (a *DeleteAction) Name() string { return "delete" }

func (a *DeleteAction) Execute(ctx *botctx.Context) error {
	err := a.api.DeleteMessage(ctx.ChatID, ctx.MessageID)
	if err != nil {
		slog.Error("failed to delete message",
			"chat_id", ctx.ChatID,
			"message_id", ctx.MessageID,
			"error", err,
		)
		return fmt.Errorf("delete action: %w", err)
	}
	slog.Info("message deleted",
		"chat_id", ctx.ChatID,
		"user_id", ctx.UserID,
		"message_id", ctx.MessageID,
	)
	return nil
}
