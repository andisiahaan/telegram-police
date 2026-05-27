package gate

import (
	"log/slog"
	"time"

	"github.com/andisiahaan/telegram-police/internal/botctx"
	"github.com/andisiahaan/telegram-police/internal/cache"
	"github.com/andisiahaan/telegram-police/internal/telegram"
)

// Gate decides whether a message should skip the entire filter pipeline.
type Gate struct {
	cacheStore *cache.Store
	api        *telegram.API
}

// New creates a Gate.
func New(cacheStore *cache.Store, api *telegram.API) *Gate {
	return &Gate{
		cacheStore: cacheStore,
		api:        api,
	}
}

// ShouldBypass returns true if the message should be left alone.
// Two conditions trigger a bypass:
//  1. The sender is in ExcludedSenderIDs
//  2. The sender is an admin or creator in this chat
func (g *Gate) ShouldBypass(ctx *botctx.Context) bool {
	if g.isExcluded(ctx) {
		ctx.Bypassed = true
		ctx.BypassReason = "User is in excluded senders list"
		return true
	}
	if g.isAdmin(ctx) {
		ctx.Bypassed = true
		ctx.BypassReason = "User is an admin or creator"
		return true
	}
	return false
}

func (g *Gate) isExcluded(ctx *botctx.Context) bool {
	for _, id := range ctx.Config.ExcludedSenderIDs {
		if id == ctx.UserID {
			return true
		}
	}
	return false
}

// isAdmin checks the cache first. On a miss it calls the Telegram API and
// stores the result. If the API call fails we treat the user as non-admin —
// fail-safe over fail-open.
func (g *Gate) isAdmin(ctx *botctx.Context) bool {
	if isAdm, found := cache.IsAdmin(g.cacheStore, ctx.ChatID, ctx.UserID); found {
		return isAdm
	}

	member, err := g.api.GetChatMember(ctx.ChatID, ctx.UserID)
	if err != nil {
		slog.Error("failed to fetch chat member status",
			"chat_id", ctx.ChatID,
			"user_id", ctx.UserID,
			"error", err,
		)
		return false
	}

	isAdm := member.Status == "administrator" || member.Status == "creator"
	ttl := time.Duration(ctx.Config.AdminCacheTTLHours) * time.Hour
	cache.SetAdmin(g.cacheStore, ctx.ChatID, ctx.UserID, isAdm, ttl)

	return isAdm
}
