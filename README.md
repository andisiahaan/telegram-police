# Telegram Police — Telegram Group Filter Bot

A self-hosted Telegram bot for filtering messages in group chats. Written in Go, webhook-based, no third-party bot libraries.

## Features

- **Dual Mode (Webhook & Poll)** — seamlessly switch between lightweight webhooks for production or long-polling for local development.
- **Auto-webhook Registration** — the bot handles its own Telegram webhook registration and deletion based on the chosen mode.
- **Modular** — each filter can be toggled on/off via config
- **No external database** — all state lives in in-memory cache
- **Dual config** — configure via `config.yaml` or environment variables (env overrides yaml)
- **Message filters**: join/left events, forwards, character length, non-latin, emoji, mentions, forbidden words, user display name
- **Rate limiting**: flood detection (sliding window) and spam violation counter
- **Actions**: delete message, ban user, or send a warning

---

## Requirements

- Go 1.22+
- *(Optional)* A public HTTPS domain if using webhook mode

---



---

## Configuration via `config.yaml`

Copy the example and fill in your values:

```bash
cp config.example.yaml config.yaml
```

Edit `config.yaml`:

```yaml
bot_token: "123456:ABC-DEF..."
bot_mode: "webhook" # or "poll"
webhook_url: "https://yourdomain.com/webhook" # auto-registered
webhook_secret: "your-secret-here"
webhook_port: 8080
```

---

## Configuration via Environment Variables

Copy the example and fill in your values:

```bash
cp .env.example .env
```

Environment variables **override** values from `config.yaml`.

### All Configuration Options

| Config YAML | Env Variable | Default | Description |
|---|---|---|---|
| `bot_token` | `BOT_TOKEN` | *(required)* | Telegram bot token |
| `bot_mode` | `BOT_MODE` | `"webhook"` | Bot operation mode (`webhook` or `poll`) |
| `webhook_url` | `WEBHOOK_URL` | `""` | Public URL for Telegram to send updates |
| `webhook_secret` | `WEBHOOK_SECRET` | `""` | Secret token for webhook validation |
| `webhook_port` | `WEBHOOK_PORT` | `8080` | HTTP server port |
| `excluded_sender_ids` | `EXCLUDED_SENDER_IDS` | `[]` | User IDs exempt from all filters (comma-separated in env) |
| `admin_cache_ttl_hours` | `ADMIN_CACHE_TTL_HOURS` | `6` | Admin status cache TTL in hours |
| `forbidden_words` | `FORBIDDEN_WORDS` | `[]` | Forbidden words in message text (comma-separated in env) |
| `forbidden_title_words` | `FORBIDDEN_TITLE_WORDS` | `[]` | Forbidden words in user display name / username |
| `max_characters` | `MAX_CHARACTERS` | `1000` | Max characters per message (0 = disabled) |
| `max_emoticons` | `MAX_EMOTICONS` | `10` | Max emoji per message (0 = disabled) |
| `max_mentions` | `MAX_MENTIONS` | `3` | Max mentions per message (0 = disabled) |
| `non_latin_threshold_percent` | `NON_LATIN_THRESHOLD_PERCENT` | `30` | % of non-latin letters before triggering the filter |
| `ban_duration_days` | `BAN_DURATION_DAYS` | `30` | Ban duration in days |
| `warn_enabled` | `WARN_ENABLED` | `true` | Send a warning message when action is warn |
| `warn_message_template` | `WARN_MESSAGE_TEMPLATE` | `"⚠️ @{username}, your message was removed: {reason}"` | Warning message template |
| `warn_auto_delete_seconds` | `WARN_AUTO_DELETE_SECONDS` | `10` | Auto-delete the warning after N seconds (0 = keep) |
| `flood.max_messages` | `FLOOD_MAX_MESSAGES` | `10` | Max messages per interval before flood action |
| `flood.interval_minutes` | `FLOOD_INTERVAL_MINUTES` | `1` | Flood detection window in minutes |
| `flood.action` | `FLOOD_ACTION` | `"ban"` | Flood action: `ban`, `delete`, or `warn` |
| `spam.max_violations` | `SPAM_MAX_VIOLATIONS` | `3` | Violations before spam action kicks in |
| `spam.window_minutes` | `SPAM_WINDOW_MINUTES` | `30` | Spam violation window in minutes |
| `spam.action` | `SPAM_ACTION` | `"ban"` | Spam action: `ban`, `delete`, or `warn` |
| `filters.delete_new_chat_members` | `FILTERS_DELETE_NEW_CHAT_MEMBERS` | `true` | Remove join notifications |
| `filters.delete_left_chat_member` | `FILTERS_DELETE_LEFT_CHAT_MEMBER` | `true` | Remove leave notifications |
| `filters.delete_forwarded_messages` | `FILTERS_DELETE_FORWARDED_MESSAGES` | `false` | Delete forwarded messages |
| `filters.check_non_latin` | `FILTERS_CHECK_NON_LATIN` | `true` | Enable non-latin character filter |

---

## Running Directly

```bash
# Install dependencies
go mod tidy

# Run
go run ./cmd/bot/
```

The bot reads `config.yaml` from the current directory, or falls back to environment variables.

---

## Running with Docker

### Build

```bash
docker build -t telegram-police .
```

### Run with config.yaml

```bash
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  --name telegram-police \
  telegram-police
```

### Run with environment variables

```bash
docker run -d \
  -p 8080:8080 \
  -e BOT_TOKEN="123456:ABC-DEF..." \
  -e WEBHOOK_SECRET="your-secret-here" \
  --name telegram-police \
  telegram-police
```

---

## License

MIT
