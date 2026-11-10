package handler

import (
	"log/slog"
	"strings"

	"github.com/andisiahaan/telegram-police/internal/botctx"
	"github.com/andisiahaan/telegram-police/internal/config"
	"github.com/andisiahaan/telegram-police/internal/pipeline"
	"github.com/andisiahaan/telegram-police/internal/telegram"
)

// Dispatcher routes incoming Telegram updates to the pipeline or handles direct interactions.
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

	// If the message is a direct chat or a status/ping command, reply indicating the bot is active.
	if d.handleDirectOrCommand(ctx) {
		return ctx
	}

	d.pl.Run(ctx)
	return ctx
}

func (d *Dispatcher) handleDirectOrCommand(ctx *botctx.Context) bool {
	if ctx.IsPrivateChat() {
		msg := "👮‍♂️ TelePolice is active!\n\nI am a group moderation bot designed to protect Telegram groups from spam, flood, and unwanted content.\n\nAdd me to your group and promote me to admin to get started!"
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

	// In group chats, handle explicit ping, status, or start commands.
	text := strings.TrimSpace(ctx.Update.Message.Text)
	if isStatusCommand(text) {
		msg := "👮‍♂️ TelePolice is active and monitoring this group."
		if _, err := d.api.SendMessage(ctx.ChatID, msg); err != nil {
			slog.Error("failed to reply to status command",
				"chat_id", ctx.ChatID,
				"user_id", ctx.UserID,
				"command", text,
				"error", err,
			)
		} else {
			slog.Info("replied to status command", "chat_id", ctx.ChatID, "command", text)
		}
		return true
	}

	return false
}

// isStatusCommand checks whether the message text starts with a command checking bot activity.
func isStatusCommand(text string) bool {
	if !strings.HasPrefix(text, "/") {
		return false
	}
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return false
	}
	cmd := strings.ToLower(fields[0])
	if atIdx := strings.Index(cmd, "@"); atIdx != -1 {
		cmd = cmd[:atIdx]
	}
	return cmd == "/ping" || cmd == "/status" || cmd == "/start"
}
