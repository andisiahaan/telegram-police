package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andisiahaan/telegram-police/internal/action"
	"github.com/andisiahaan/telegram-police/internal/cache"
	"github.com/andisiahaan/telegram-police/internal/config"
	"github.com/andisiahaan/telegram-police/internal/filter"
	"github.com/andisiahaan/telegram-police/internal/gate"
	"github.com/andisiahaan/telegram-police/internal/handler"
	"github.com/andisiahaan/telegram-police/internal/pipeline"
	"github.com/andisiahaan/telegram-police/internal/ratelimit"
	"github.com/andisiahaan/telegram-police/internal/server"
	"github.com/andisiahaan/telegram-police/internal/telegram"
)

func main() {
	// 1. Load config
	cfg, err := config.Load("config.yaml")
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// 2. Core dependencies
	cacheStore := cache.New(
		time.Duration(cfg.AdminCacheTTLHours)*time.Hour,
		10*time.Minute,
	)
	tgClient := telegram.NewClient(cfg.BotToken)
	tgAPI := telegram.NewAPI(tgClient)

	// 3. Rate limiters
	floodChecker := ratelimit.NewFloodChecker(cfg)
	spamChecker := ratelimit.NewSpamChecker(cfg)

	// 4. Filter chain
	filters := []filter.Filter{
		&filter.NewChatMembersFilter{},
		&filter.LeftChatMemberFilter{},
		&filter.ForwardedMessageFilter{},
		&filter.ContentLengthFilter{},
		&filter.NonLatinFilter{},
		&filter.EmoticonFilter{},
		&filter.MentionFilter{},
		&filter.ForbiddenWordsFilter{},
		&filter.UserTitleFilter{},
	}

	// 5. Gate, resolver, pipeline
	g := gate.New(cacheStore, tgAPI)
	resolver := action.NewResolver(tgAPI, spamChecker)
	pl := pipeline.New(g, floodChecker, filters, resolver, cfg)

	// 6. HTTP handler
	dispatcher := handler.NewDispatcher(pl, cfg)
	webhookHandler := handler.NewWebhookHandler(dispatcher, cfg.WebhookSecret)

	// 7. Server + graceful shutdown on SIGINT / SIGTERM
	srv := server.New(cfg.WebhookPort, webhookHandler)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		srv.Shutdown()
	}()

	if err := srv.Start(); err != nil {
		slog.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}
