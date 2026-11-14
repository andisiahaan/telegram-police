package handler

import (
	"testing"

	"github.com/andisiahaan/telegram-police/internal/botctx"
)

func TestHandlePrivateChatNonPrivate(t *testing.T) {
	d := &Dispatcher{}

	tests := []struct {
		name     string
		chatType string
	}{
		{"group chat", "group"},
		{"supergroup chat", "supergroup"},
		{"channel", "channel"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &botctx.Context{
				ChatType: tt.chatType,
			}
			handled := d.handlePrivateChat(ctx)
			if handled {
				t.Errorf("handlePrivateChat should return false for %s, got true", tt.chatType)
			}
		})
	}
}
