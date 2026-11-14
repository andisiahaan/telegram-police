package config

import "testing"

func TestURLResolvers(t *testing.T) {
	tests := []struct {
		name            string
		appURL          string
		expectedWebhook string
		expectedAdmin   string
	}{
		{
			name:            "standard url",
			appURL:          "https://telepolice.com",
			expectedWebhook: "https://telepolice.com/webhook",
			expectedAdmin:   "https://telepolice.com/admin",
		},
		{
			name:            "url with trailing slash",
			appURL:          "https://telepolice.com/",
			expectedWebhook: "https://telepolice.com/webhook",
			expectedAdmin:   "https://telepolice.com/admin",
		},
		{
			name:            "url with /webhook suffix",
			appURL:          "https://telepolice.com/webhook",
			expectedWebhook: "https://telepolice.com/webhook",
			expectedAdmin:   "https://telepolice.com/admin",
		},
		{
			name:            "url with /admin suffix",
			appURL:          "https://telepolice.com/admin",
			expectedWebhook: "https://telepolice.com/webhook",
			expectedAdmin:   "https://telepolice.com/admin",
		},
		{
			name:            "empty url",
			appURL:          "",
			expectedWebhook: "",
			expectedAdmin:   "/admin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{AppURL: tt.appURL}
			if got := cfg.WebhookURL(); got != tt.expectedWebhook {
				t.Errorf("WebhookURL() = %q, want %q", got, tt.expectedWebhook)
			}
			if got := cfg.AdminURL(); got != tt.expectedAdmin {
				t.Errorf("AdminURL() = %q, want %q", got, tt.expectedAdmin)
			}
		})
	}
}
