package filter

import (
	"testing"

	"github.com/andisiahaan/telegram-police/internal/botctx"
	"github.com/andisiahaan/telegram-police/internal/config"
	"github.com/andisiahaan/telegram-police/internal/telegram"
)

func TestForbiddenWordsFilter(t *testing.T) {
	cfg := &config.Config{
		ForbiddenWords: []string{"badword", "c++"},
	}

	filter := NewForbiddenWordsFilter(cfg.ForbiddenWords)

	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{"exact match", "this is a badword", true},
		{"uppercase match", "this is a BADWORD", true},
		{"no match", "this is a goodword", false},
		{"partial word match should not trigger", "this is a badwordless sentence", false},
		{"special characters", "I like c++ programming", true},
		{"special character partial", "I like c++less", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &botctx.Context{
				Update: &telegram.Update{
					Message: &telegram.Message{
						Text: tt.text,
					},
				},
				Config: cfg,
			}
			violated, _ := filter.Check(ctx)
			if violated != tt.expected {
				t.Errorf("expected %v, got %v for text: %s", tt.expected, violated, tt.text)
			}
		})
	}
}
