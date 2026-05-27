package handler

import (
	"github.com/andisiahaan/telegram-police/internal/botctx"
	"github.com/andisiahaan/telegram-police/internal/config"
	"github.com/andisiahaan/telegram-police/internal/pipeline"
	"github.com/andisiahaan/telegram-police/internal/telegram"
)

// Dispatcher routes incoming Telegram updates to the pipeline.
type Dispatcher struct {
	pl  *pipeline.Pipeline
	cfg *config.Config
}

// NewDispatcher creates a Dispatcher.
func NewDispatcher(pl *pipeline.Pipeline, cfg *config.Config) *Dispatcher {
	return &Dispatcher{pl: pl, cfg: cfg}
}

// Dispatch processes a single Telegram Update. Only updates with a Message field
// are handled; edited messages, channel posts, etc. are ignored for now.
func (d *Dispatcher) Dispatch(update *telegram.Update) *botctx.Context {
	if update.Message == nil {
		return nil
	}
	ctx := botctx.New(update, d.cfg)
	d.pl.Run(ctx)
	return ctx
}
