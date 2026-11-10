package handler

import (
	"testing"
)

func TestIsStatusCommand(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"/ping", true},
		{"/ping@telepolice_bot", true},
		{"/status", true},
		{"/status@telepolice_bot", true},
		{"/start", true},
		{"/start@telepolice_bot", true},
		{"/help", false},
		{"halo", false},
		{"ping", false},
		{"", false},
	}

	for _, tt := range tests {
		got := isStatusCommand(tt.input)
		if got != tt.expected {
			t.Errorf("isStatusCommand(%q) = %v; want %v", tt.input, got, tt.expected)
		}
	}
}

