package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
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

	slog.Info("starting telepolice", "mode", cfg.BotMode)

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
		filter.NewForbiddenWordsFilter(cfg.ForbiddenWords),
		&filter.UserTitleFilter{},
	}

	// 5. WaitGroup for background tasks
	var wg sync.WaitGroup

	// 6. Gate, resolver, pipeline, dispatcher
	g := gate.New(cacheStore, tgAPI)
	resolver := action.NewResolver(tgAPI, spamChecker, &wg)
	pl := pipeline.New(g, floodChecker, filters, resolver, cfg)
	dispatcher := handler.NewDispatcher(pl, cfg)

	// 7. Listen for OS shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	switch cfg.BotMode {
	case "poll":
		slog.Info("removing webhook (if any) to enable polling")
		if err := tgAPI.DeleteWebhook(); err != nil {
			slog.Warn("failed to delete webhook (safe to ignore if not set)", "error", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			<-quit
			cancel()
			wg.Wait()
			slog.Info("all background tasks finished")
		}()
		handler.NewPoller(dispatcher, tgAPI, &wg).Start(ctx)

	default: // "webhook"
		if cfg.WebhookURL != "" {
			slog.Info("registering webhook", "url", cfg.WebhookURL)
			if err := tgAPI.SetWebhook(cfg.WebhookURL, cfg.WebhookSecret); err != nil {
				slog.Error("failed to set webhook", "error", err)
				os.Exit(1)
			}
		}

		webhookHandler := handler.NewWebhookHandler(dispatcher, cfg.WebhookSecret, &wg)
		srv := server.New(cfg.WebhookPort, webhookHandler)
		go func() {
			<-quit
			srv.Shutdown()
			wg.Wait()
			slog.Info("all background tasks finished")
		}()
		if err := srv.Start(); err != nil {
			slog.Error("server exited with error", "error", err)
			os.Exit(1)
		}
	}
}
