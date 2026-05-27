package action

import (
	"github.com/andisiahaan/telegram-police/internal/botctx"
)

// Action is the interface that all bot actions must implement.
type Action interface {
	Name() string
	Execute(ctx *botctx.Context) error
}
