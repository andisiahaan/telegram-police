package config

// applyDefaults fills in zero-value fields with sensible defaults.
func applyDefaults(cfg *Config) {
	if cfg.WebhookPort == 0 {
		cfg.WebhookPort = 8080
	}
	if cfg.AdminCacheTTLHours == 0 {
		cfg.AdminCacheTTLHours = 6
	}
	if cfg.MaxCharacters == 0 {
		cfg.MaxCharacters = 1000
	}
	if cfg.MaxEmoticons == 0 {
		cfg.MaxEmoticons = 10
	}
	if cfg.MaxMentions == 0 {
		cfg.MaxMentions = 3
	}
	if cfg.NonLatinThresholdPercent == 0 {
		cfg.NonLatinThresholdPercent = 30
	}
	if cfg.BanDurationDays == 0 {
		cfg.BanDurationDays = 30
	}
	if cfg.WarnMessageTemplate == "" {
		cfg.WarnMessageTemplate = "⚠️ @{username}, your message was removed: {reason}"
	}
	if cfg.WarnAutoDeleteSeconds == 0 {
		cfg.WarnAutoDeleteSeconds = 10
	}
	applyFloodDefaults(&cfg.Flood)
	applySpamDefaults(&cfg.Spam)
	applyFiltersDefaults(&cfg.Filters)
}

func applyFloodDefaults(f *FloodConfig) {
	if f.MaxMessages == 0 {
		f.MaxMessages = 10
	}
	if f.IntervalMinutes == 0 {
		f.IntervalMinutes = 1
	}
	if f.Action == "" {
		f.Action = "ban"
	}
}

func applySpamDefaults(s *SpamConfig) {
	if s.MaxViolations == 0 {
		s.MaxViolations = 3
	}
	if s.WindowMinutes == 0 {
		s.WindowMinutes = 30
	}
	if s.Action == "" {
		s.Action = "ban"
	}
}

func applyFiltersDefaults(f *FiltersConfig) {
	// Bool zero-value is false, so the true defaults (delete_new_chat_members,
	// delete_left_chat_member, check_non_latin) are set in Load() before YAML
	// is parsed, not here.
	_ = f
}
