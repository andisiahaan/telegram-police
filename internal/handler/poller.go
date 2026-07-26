package handler

import (
	"context"
	"log/slog"
	"sync"

	"github.com/andisiahaan/telegram-police/internal/telegram"
)

// Poller runs a long-polling loop against the Telegram getUpdates API.
type Poller struct {
	dispatcher *Dispatcher
	api        *telegram.API
	wg         *sync.WaitGroup
}

// NewPoller creates a Poller.
func NewPoller(dispatcher *Dispatcher, api *telegram.API, wg *sync.WaitGroup) *Poller {
	return &Poller{dispatcher: dispatcher, api: api, wg: wg}
}

// Start runs the polling loop until ctx is cancelled.
// Each update is dispatched in a goroutine and tracked by the WaitGroup.
func (p *Poller) Start(ctx context.Context) {
	slog.Info("poller: starting long-poll mode")
	offset := 0
	const pollTimeoutSec = 30

	for {
		select {
		case <-ctx.Done():
			slog.Info("poller: context cancelled, stopping")
			return
		default:
		}

		updates, err := p.api.GetUpdates(offset, pollTimeoutSec)
		if err != nil {
			slog.Error("poller: getUpdates failed", "error", err)
			continue
		}

		for _, u := range updates {
			offset = u.UpdateID + 1
			update := u // capture loop variable
			if p.wg != nil {
				p.wg.Add(1)
			}
			go func() {
				if p.wg != nil {
					defer p.wg.Done()
				}
				p.dispatcher.Dispatch(&update)
			}()
		}
	}
}
