package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Config holds the full bot configuration.
type Config struct {
	BotToken      string `yaml:"bot_token"`
	WebhookSecret string `yaml:"webhook_secret"`
	WebhookPort   int    `yaml:"webhook_port"`

	ExcludedSenderIDs        []int64  `yaml:"excluded_sender_ids"`
	AdminCacheTTLHours       int      `yaml:"admin_cache_ttl_hours"`
	ForbiddenWords           []string `yaml:"forbidden_words"`
	ForbiddenTitleWords      []string `yaml:"forbidden_title_words"`
	MaxCharacters            int      `yaml:"max_characters"`
	MaxEmoticons             int      `yaml:"max_emoticons"`
	MaxMentions              int      `yaml:"max_mentions"`
	NonLatinThresholdPercent int      `yaml:"non_latin_threshold_percent"`
	BanDurationDays          int      `yaml:"ban_duration_days"`

	WarnEnabled           bool   `yaml:"warn_enabled"`
	WarnMessageTemplate   string `yaml:"warn_message_template"`
	WarnAutoDeleteSeconds int    `yaml:"warn_auto_delete_seconds"`

	Flood   FloodConfig   `yaml:"flood"`
	Spam    SpamConfig    `yaml:"spam"`
	Filters FiltersConfig `yaml:"filters"`
}

// FloodConfig controls the flood detection behaviour.
type FloodConfig struct {
	MaxMessages     int    `yaml:"max_messages"`
	IntervalMinutes int    `yaml:"interval_minutes"`
	Action          string `yaml:"action"`
}

// SpamConfig controls the accumulated-violations (spam) behaviour.
type SpamConfig struct {
	MaxViolations int    `yaml:"max_violations"`
	WindowMinutes int    `yaml:"window_minutes"`
	Action        string `yaml:"action"`
}

// FiltersConfig lets you toggle each filter on or off.
type FiltersConfig struct {
	DeleteNewChatMembers    bool `yaml:"delete_new_chat_members"`
	DeleteLeftChatMember    bool `yaml:"delete_left_chat_member"`
	DeleteForwardedMessages bool `yaml:"delete_forwarded_messages"`
	CheckNonLatin           bool `yaml:"check_non_latin"`
}

// Load reads config from a YAML file and then overrides any field with an
// environment variable. Priority: defaults → config.yaml → .env → system env.
func Load(path string) (*Config, error) {
	// Load .env if present; ignore the error if the file doesn't exist.
	_ = godotenv.Load()

	cfg := &Config{}

	// Set bool filters that default to true before YAML parsing overwrites them.
	cfg.Filters.DeleteNewChatMembers = true
	cfg.Filters.DeleteLeftChatMember = true
	cfg.Filters.CheckNonLatin = true
	cfg.WarnEnabled = true

	applyDefaults(cfg)

	// Parse YAML — optional, no error if the file is missing.
	if data, err := os.ReadFile(path); err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("config: parse yaml: %w", err)
		}
	}

	overrideFromEnv(cfg)

	if cfg.BotToken == "" {
		return nil, fmt.Errorf("config: BOT_TOKEN or bot_token in config.yaml is required")
	}

	return cfg, nil
}

// overrideFromEnv applies environment variable values on top of whatever was
// loaded from the YAML file. Only non-empty env values are applied.
func overrideFromEnv(cfg *Config) {
	if v := os.Getenv("BOT_TOKEN"); v != "" {
		cfg.BotToken = v
	}
	if v := os.Getenv("WEBHOOK_SECRET"); v != "" {
		cfg.WebhookSecret = v
	}
	if v := os.Getenv("WEBHOOK_PORT"); v != "" {
		cfg.WebhookPort = parseInt(v, cfg.WebhookPort)
	}
	if v := os.Getenv("EXCLUDED_SENDER_IDS"); v != "" {
		cfg.ExcludedSenderIDs = parseInt64Slice(v)
	}
	if v := os.Getenv("ADMIN_CACHE_TTL_HOURS"); v != "" {
		cfg.AdminCacheTTLHours = parseInt(v, cfg.AdminCacheTTLHours)
	}
	if v := os.Getenv("FORBIDDEN_WORDS"); v != "" {
		cfg.ForbiddenWords = parseStringSlice(v)
	}
	if v := os.Getenv("FORBIDDEN_TITLE_WORDS"); v != "" {
		cfg.ForbiddenTitleWords = parseStringSlice(v)
	}
	if v := os.Getenv("MAX_CHARACTERS"); v != "" {
		cfg.MaxCharacters = parseInt(v, cfg.MaxCharacters)
	}
	if v := os.Getenv("MAX_EMOTICONS"); v != "" {
		cfg.MaxEmoticons = parseInt(v, cfg.MaxEmoticons)
	}
	if v := os.Getenv("MAX_MENTIONS"); v != "" {
		cfg.MaxMentions = parseInt(v, cfg.MaxMentions)
	}
	if v := os.Getenv("NON_LATIN_THRESHOLD_PERCENT"); v != "" {
		cfg.NonLatinThresholdPercent = parseInt(v, cfg.NonLatinThresholdPercent)
	}
	if v := os.Getenv("BAN_DURATION_DAYS"); v != "" {
		cfg.BanDurationDays = parseInt(v, cfg.BanDurationDays)
	}
	if v := os.Getenv("WARN_ENABLED"); v != "" {
		cfg.WarnEnabled = parseBool(v, cfg.WarnEnabled)
	}
	if v := os.Getenv("WARN_MESSAGE_TEMPLATE"); v != "" {
		cfg.WarnMessageTemplate = v
	}
	if v := os.Getenv("WARN_AUTO_DELETE_SECONDS"); v != "" {
		cfg.WarnAutoDeleteSeconds = parseInt(v, cfg.WarnAutoDeleteSeconds)
	}
	if v := os.Getenv("FLOOD_MAX_MESSAGES"); v != "" {
		cfg.Flood.MaxMessages = parseInt(v, cfg.Flood.MaxMessages)
	}
	if v := os.Getenv("FLOOD_INTERVAL_MINUTES"); v != "" {
		cfg.Flood.IntervalMinutes = parseInt(v, cfg.Flood.IntervalMinutes)
	}
	if v := os.Getenv("FLOOD_ACTION"); v != "" {
		cfg.Flood.Action = v
	}
	if v := os.Getenv("SPAM_MAX_VIOLATIONS"); v != "" {
		cfg.Spam.MaxViolations = parseInt(v, cfg.Spam.MaxViolations)
	}
	if v := os.Getenv("SPAM_WINDOW_MINUTES"); v != "" {
		cfg.Spam.WindowMinutes = parseInt(v, cfg.Spam.WindowMinutes)
	}
	if v := os.Getenv("SPAM_ACTION"); v != "" {
		cfg.Spam.Action = v
	}
	if v := os.Getenv("FILTERS_DELETE_NEW_CHAT_MEMBERS"); v != "" {
		cfg.Filters.DeleteNewChatMembers = parseBool(v, cfg.Filters.DeleteNewChatMembers)
	}
	if v := os.Getenv("FILTERS_DELETE_LEFT_CHAT_MEMBER"); v != "" {
		cfg.Filters.DeleteLeftChatMember = parseBool(v, cfg.Filters.DeleteLeftChatMember)
	}
	if v := os.Getenv("FILTERS_DELETE_FORWARDED_MESSAGES"); v != "" {
		cfg.Filters.DeleteForwardedMessages = parseBool(v, cfg.Filters.DeleteForwardedMessages)
	}
	if v := os.Getenv("FILTERS_CHECK_NON_LATIN"); v != "" {
		cfg.Filters.CheckNonLatin = parseBool(v, cfg.Filters.CheckNonLatin)
	}
}

func parseInt(s string, fallback int) int {
	v, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return fallback
	}
	return v
}

func parseBool(s string, fallback bool) bool {
	v, err := strconv.ParseBool(strings.TrimSpace(s))
	if err != nil {
		return fallback
	}
	return v
}

func parseStringSlice(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			result = append(result, t)
		}
	}
	return result
}

func parseInt64Slice(s string) []int64 {
	parts := strings.Split(s, ",")
	result := make([]int64, 0, len(parts))
	for _, p := range parts {
		if v, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64); err == nil {
			result = append(result, v)
		}
	}
	return result
}
