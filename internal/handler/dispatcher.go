package handler

import (
	"log/slog"

	"github.com/andisiahaan/telegram-police/internal/botctx"
	"github.com/andisiahaan/telegram-police/internal/config"
	"github.com/andisiahaan/telegram-police/internal/pipeline"
	"github.com/andisiahaan/telegram-police/internal/telegram"
)

// Dispatcher routes incoming Telegram updates to the pipeline or handles private chat interactions.
type Dispatcher struct {
	pl  *pipeline.Pipeline
	api *telegram.API
	cfg *config.Config
}

// NewDispatcher creates a Dispatcher.
func NewDispatcher(pl *pipeline.Pipeline, api *telegram.API, cfg *config.Config) *Dispatcher {
	return &Dispatcher{
		pl:  pl,
		api: api,
		cfg: cfg,
	}
}

// Dispatch processes a single Telegram Update. Only updates with a Message field
// are handled; edited messages, channel posts, etc. are ignored for now.
func (d *Dispatcher) Dispatch(update *telegram.Update) *botctx.Context {
	if update.Message == nil {
		return nil
	}
	ctx := botctx.New(update, d.cfg)

	// If the message is a direct 1-on-1 private chat, reply explaining that the bot can manage groups.
	if d.handlePrivateChat(ctx) {
		return ctx
	}

	d.pl.Run(ctx)
	return ctx
}

func (d *Dispatcher) handlePrivateChat(ctx *botctx.Context) bool {
	if !ctx.IsPrivateChat() {
		return false
	}

	msg := "👮‍♂️ TelePolice is active!\n\nI am a group moderation bot designed to protect and manage Telegram groups from spam, flood, and unwanted content.\n\nAdd me to your group and promote me to admin to manage the group!"
	if _, err := d.api.SendMessage(ctx.ChatID, msg); err != nil {
		slog.Error("failed to reply to private chat",
			"chat_id", ctx.ChatID,
			"user_id", ctx.UserID,
			"error", err,
		)
	} else {
		slog.Info("replied to private chat", "chat_id", ctx.ChatID, "user_id", ctx.UserID)
	}
	return true
}
