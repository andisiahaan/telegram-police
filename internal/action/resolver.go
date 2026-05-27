package action

import (
	"log/slog"

	"github.com/andisiahaan/telegram-police/internal/botctx"
	"github.com/andisiahaan/telegram-police/internal/ratelimit"
	"github.com/andisiahaan/telegram-police/internal/telegram"
)

// Resolver decides which actions to run based on the pipeline context.
type Resolver struct {
	deleteAction *DeleteAction
	banAction    *BanAction
	warnAction   *WarnAction
	spamChecker  *ratelimit.SpamChecker
}

// NewResolver creates a Resolver with all required dependencies.
func NewResolver(api *telegram.API, spamChecker *ratelimit.SpamChecker) *Resolver {
	return &Resolver{
		deleteAction: NewDeleteAction(api),
		banAction:    NewBanAction(api),
		warnAction:   NewWarnAction(api),
		spamChecker:  spamChecker,
	}
}

// Resolve handles a message that triggered one or more filter violations.
// It always deletes the message first, then checks whether the spam threshold
// has been hit and fires the configured spam action if so.
func (r *Resolver) Resolve(ctx *botctx.Context) {
	if !ctx.HasViolations() {
		return
	}

	r.execute(r.deleteAction, ctx)

	r.spamChecker.AddViolation(ctx)
	if r.spamChecker.Check(ctx) {
		action := r.resolveActionByName(ctx.Config.Spam.Action)
		r.execute(action, ctx)
	}
}

// ResolveFlood fires the configured flood action. Called by the pipeline when
// the flood threshold is exceeded.
func (r *Resolver) ResolveFlood(ctx *botctx.Context) {
	action := r.resolveActionByName(ctx.Config.Flood.Action)
	r.execute(action, ctx)
}

func (r *Resolver) resolveActionByName(name string) Action {
	switch name {
	case "warn":
		return r.warnAction
	case "ban":
		return r.banAction
	default:
		return r.deleteAction
	}
}

func (r *Resolver) execute(a Action, ctx *botctx.Context) {
	if err := a.Execute(ctx); err != nil {
		slog.Error("action failed",
			"action", a.Name(),
			"chat_id", ctx.ChatID,
			"user_id", ctx.UserID,
			"error", err,
		)
	} else {
		ctx.AddAction(a.Name())
	}
}
