# Blueprint: Telegram Group Filter Bot (Go)

## Konteks Proyek

Buat Telegram Bot Filter dengan GO
Akan di-open source sebagai self-hosted bot, akan dijadikan repo public di github
Prioritas: modular, SSOT, DRY, mudah di-deploy siapapun tanpa dependensi eksternal.

---

## Prinsip Wajib

- Setiap file maksimal 200–300 baris. Jika melebihi, wajib dipecah.
- Tidak boleh ada logika bisnis di `main.go`. Hanya wiring/DI.
- Tidak ada Redis atau database eksternal. Semua state pakai in-memory cache.
- Semua konfigurasi bisa diisi via `config.yaml` ATAU environment variable (dua-duanya didukung, env override yaml).
- Webhook mode, bukan polling.
- Setiap filter adalah unit independen yang bisa diaktifkan/dinonaktifkan via config.
- Tidak pakai library bot Telegram pihak ketiga. Semua panggilan ke Telegram API dilakukan via HTTP client sendiri.

---

## Struktur Folder

```
telegram-filter-bot/
├── cmd/
│   └── bot/
│       └── main.go
├── internal/
│   ├── config/
│   │   ├── config.go
│   │   └── defaults.go
│   ├── server/
│   │   └── server.go
│   ├── handler/
│   │   ├── webhook.go
│   │   └── dispatcher.go
│   ├── pipeline/
│   │   ├── pipeline.go
│   │   └── context.go
│   ├── gate/
│   │   └── gate.go
│   ├── filter/
│   │   ├── filter.go
│   │   ├── event_join.go
│   │   ├── event_left.go
│   │   ├── event_forward.go
│   │   ├── content_length.go
│   │   ├── content_nonlatin.go
│   │   ├── content_emoticon.go
│   │   ├── content_mention.go
│   │   ├── content_words.go
│   │   └── user_title.go
│   ├── ratelimit/
│   │   ├── window.go
│   │   ├── flood.go
│   │   └── spam.go
│   ├── action/
│   │   ├── action.go
│   │   ├── delete.go
│   │   ├── ban.go
│   │   └── resolver.go
│   ├── cache/
│   │   ├── store.go
│   │   └── admin.go
│   └── telegram/
│       ├── types.go
│       ├── client.go
│       └── api.go
├── config.example.yaml
├── .env.example
├── Dockerfile
├── go.mod
└── README.md
```

---

## Dependensi Eksternal (go.mod)

| Package | Kegunaan |
|---|---|
| `gopkg.in/yaml.v3` | Parse config.yaml |
| `github.com/patrickmn/go-cache` | In-memory cache dengan TTL |
| `github.com/joho/godotenv` | Load file .env |

Tidak ada dependensi lain. Semua networking pakai `net/http` standar.

---

## Spesifikasi Konfigurasi

### Struktur `config.example.yaml`

```yaml
bot_token: ""
webhook_secret: ""  (optional)
webhook_port: 8080

excluded_sender_ids: []
admin_cache_ttl_hours: 6

forbidden_words: []
forbidden_title_words: []

max_characters: 1000
max_emoticons: 10
max_mentions: 3

flood:
  max_messages: 10
  interval_minutes: 1
  action: "ban"        # nilai: ban | delete | warn (warn = hanya hapus pesan)

spam:
  max_violations: 3
  window_minutes: 30
  action: "ban"

ban_duration_days: 30

filters:
  delete_new_chat_members: true
  delete_left_chat_member: true
  delete_forwarded_messages: false
  check_non_latin: true
```

Semua field di atas harus bisa di-override dengan environment variable.
Konvensi penamaan env: `BOT_TOKEN`, `WEBHOOK_SECRET`, `FLOOD_ACTION`, `FILTERS_CHECK_NON_LATIN`, dst.

---

## Arsitektur Alur Eksekusi

```
[Telegram] → POST /webhook
    │
    ▼
[handler/webhook.go]
  - Validasi X-Telegram-Bot-Api-Secret-Token header
  - Parse body JSON menjadi struct Update
    │
    ▼
[handler/dispatcher.go]
  - Identifikasi tipe update (message / edited_message / dll)
  - Teruskan ke pipeline
    │
    ▼
[pipeline/pipeline.go]   ← orchestrator utama
    │
    ├─ [gate/gate.go]
    │    - Cek apakah sender ada di excluded_sender_ids → bypass semua
    │    - Cek apakah sender adalah admin/creator di chat (via cache) → bypass semua
    │    - Jika bypass: return, tidak ada aksi
    │
    ├─ [ratelimit/flood.go]
    │    - Tambahkan pesan ke sliding window user ini
    │    - Jika melampaui max_messages_per_interval:
    │        → jalankan aksi sesuai flood.action di config
    │        → return, hentikan pipeline
    │
    ├─ [filter/filter.go] ← jalankan semua filter yang aktif secara berurutan
    │    - Kumpulkan semua violations (jangan berhenti di pelanggaran pertama)
    │    - Setiap filter mengembalikan: violated bool + reason string
    │
    └─ [action/resolver.go]
         - Jika ada violations:
             → selalu jalankan action DELETE pesan
             → tambahkan 1 ke spam counter user ini (ratelimit/spam.go)
             → jika spam counter melampaui max_spam_count:
                 → jalankan aksi sesuai spam.action di config
```

---

## Spesifikasi Tiap Modul

### `internal/config/`

**config.go**
- Struct `Config` yang merepresentasikan seluruh config.yaml
- Fungsi `Load()` yang membaca config.yaml lalu di-override oleh env variable
- Validasi minimal: bot_token tidak boleh kosong

**defaults.go**
- Nilai default untuk semua field config jika tidak diisi

---

### `internal/server/`

**server.go**
- Setup `net/http` server
- Register satu route: `POST /webhook`
- Graceful shutdown saat menerima sinyal OS (SIGINT, SIGTERM)
- Port diambil dari config

---

### `internal/handler/`

**webhook.go**
- Validasi header `X-Telegram-Bot-Api-Secret-Token`
- Baca body request, decode JSON ke struct `telegram.Update`
- Teruskan ke dispatcher
- Selalu respond `200 OK` ke Telegram secepat mungkin (proses di goroutine terpisah)

**dispatcher.go**
- Terima `telegram.Update`
- Routing: hanya proses update yang memiliki field `Message` (abaikan `EditedMessage`, `ChannelPost`, dll — kecuali ada kebutuhan di masa depan)
- Panggil `pipeline.Run(ctx)`

---

### `internal/pipeline/`

**context.go**
- Struct `Context` berisi:
  - `Update` (raw update dari Telegram)
  - `Config` (pointer ke config global)
  - `ChatID`, `UserID`, `MessageID` (shortcut)
  - `Violations` (slice of string, dikumpulkan dari filter)

**pipeline.go**
- Fungsi `Run(ctx *Context)`
- Orkestrasi urutan: Gate → Flood → FilterChain → Resolver
- Tidak boleh ada logika bisnis di sini, hanya memanggil modul lain

---

### `internal/gate/`

**gate.go**
- Fungsi `ShouldBypass(ctx *Context) bool`
- Cek 1: apakah `UserID` ada di `config.ExcludedSenderIDs`
- Cek 2: panggil `cache/admin.go` untuk cek apakah user adalah admin/creator
- Jika cache miss: panggil `telegram/api.go` → `getChatMember`, simpan hasilnya ke cache dengan TTL dari `admin_cache_ttl_hours`

---

### `internal/filter/`

**filter.go**
- Definisi interface `Filter`:
  - `Enabled(cfg *Config) bool` — apakah filter ini aktif berdasarkan config
  - `Name() string` — nama filter untuk logging/reason
  - `Check(ctx *Context) (violated bool, reason string)`
- Fungsi `RunChain(filters []Filter, ctx *Context)` yang menjalankan semua filter, mengumpulkan semua violations ke `ctx.Violations`

**event_join.go**
- Aktif jika `filters.delete_new_chat_members: true`
- Periksa apakah `Message.NewChatMembers` tidak nil/kosong

**event_left.go**
- Aktif jika `filters.delete_left_chat_member: true`
- Periksa apakah `Message.LeftChatMember` tidak nil

**event_forward.go**
- Aktif jika `filters.delete_forwarded_messages: true`
- Periksa apakah `Message.ForwardFrom` atau `Message.ForwardFromChat` tidak nil

**content_length.go**
- Aktif jika `max_characters > 0`
- Hitung panjang karakter teks pesan
- Violated jika melebihi `max_characters`

**content_nonlatin.go**
- Aktif jika `filters.check_non_latin: true`
- Deteksi karakter non-latin menggunakan package `unicode` standar Go
- Target skrip: Arabic, Cyrillic, CJK, Devanagari, Thai, Hebrew, dan skrip non-latin lainnya
- Violated jika ditemukan karakter dari skrip tersebut

**content_emoticon.go**
- Aktif jika `max_emoticons > 0`
- Hitung jumlah emoji/emoticon dalam teks menggunakan unicode range
- Violated jika melebihi `max_emoticons`

**content_mention.go**
- Aktif jika `max_mentions > 0`
- Hitung jumlah `@mention` dalam teks pesan
- Gunakan `Message.Entities` dari Telegram jika tersedia, fallback ke string matching
- Violated jika melebihi `max_mentions`

**content_words.go**
- Aktif jika `forbidden_words` tidak kosong
- Cek apakah teks pesan (lowercase) mengandung salah satu kata dari `forbidden_words`
- Case-insensitive matching

**user_title.go**
- Aktif jika `forbidden_title_words` tidak kosong
- Cek field: `User.FirstName`, `User.LastName`, `User.Username`
- Case-insensitive matching terhadap `forbidden_title_words`

---

### `internal/ratelimit/`

**window.go**
- Implementasi Sliding Window generic yang thread-safe
- Struct `SlidingWindow` dengan field: timestamps (slice of time.Time), mutex, max count, interval duration
- Method: `Add(key string)`, `Count(key string) int`, `Exceeded(key string) bool`
- Key format: `"chatID:userID"` untuk isolasi per user per grup
- Pembersihan otomatis timestamp yang sudah di luar window

**flood.go**
- Menggunakan `window.go`
- Fungsi `CheckFlood(ctx *Context) (exceeded bool)`
- Window config: `flood.max_messages` dan `flood.interval_minutes`

**spam.go**
- Menggunakan `window.go`
- Fungsi `AddViolation(ctx *Context)` dan `CheckSpam(ctx *Context) (exceeded bool)`
- Window config: `spam.max_violations` dan `spam.window_minutes`

---

### `internal/action/`

**action.go**
- Definisi interface `Action`:
  - `Name() string`
  - `Execute(ctx *Context) error`

**delete.go**
- Implementasi Action: panggil `telegram/api.go` → `DeleteMessage`

**ban.go**
- Implementasi Action: panggil `telegram/api.go` → `BanChatMember`
- Durasi ban dihitung dari `ban_duration_days` → dikonversi ke unix timestamp untuk Telegram API

**resolver.go**
- Fungsi `Resolve(ctx *Context) []Action`
- Logika:
  - Jika `ctx.Violations` tidak kosong → tambahkan action `Delete`
  - Panggil `spam.AddViolation` → jika `spam.CheckSpam` exceeded → tambahkan action sesuai `spam.action` di config
  - Untuk flood: action ditentukan sesuai `flood.action` di config
- Mapping string action ke struct Action: `"ban"` → `ban.Action{}`, `"delete"` → `delete.Action{}`, `"warn"` → hanya delete pesan tanpa penalti tambahan

---

### `internal/cache/`

**store.go**
- Wrapper tipis di atas `patrickmn/go-cache`
- Expose method: `Set(key, value, ttl)`, `Get(key)`, `Delete(key)`
- Satu instance global, diinisialisasi di `main.go` dan diinjeksikan

**admin.go**
- Fungsi `IsAdmin(chatID, userID int64) (bool, bool)` — return (isAdmin, found)
- Fungsi `SetAdmin(chatID, userID int64, isAdmin bool, ttl duration)`
- Key format: `"admin:chatID:userID"`

---

### `internal/telegram/`

**types.go**
- Definisi semua struct yang dibutuhkan dari Telegram API:
  - `Update`, `Message`, `User`, `Chat`, `MessageEntity`
  - `ChatMember` (untuk getChatMember response)
  - Field yang dibutuhkan saja, tidak perlu semua field Telegram

**client.go**
- HTTP client wrapper dengan:
  - Timeout yang wajar (10 detik)
  - Base URL: `https://api.telegram.org/bot{token}/`
  - Fungsi helper untuk POST JSON dan decode response
  - Logging error tanpa panic

**api.go**
- Fungsi-fungsi pemanggilan Telegram API:
  - `DeleteMessage(chatID int64, messageID int) error`
  - `BanChatMember(chatID, userID int64, untilDate int64) error`
  - `GetChatMember(chatID, userID int64) (*ChatMember, error)`
- Setiap fungsi menangani error response dari Telegram (field `ok: false`)

---

### `cmd/bot/main.go`

Hanya berisi:
1. Load config
2. Inisialisasi semua dependensi (cache store, telegram client, sliding windows)
3. Inject dependensi ke semua modul yang membutuhkan
4. Jalankan server
5. Tidak ada logika bisnis sama sekali

---

## Dockerfile

- Base image: `golang:alpine` untuk build, `alpine` untuk runtime
- Multi-stage build
- Binary disimpan di `/app/bot`
- Config dibaca dari `/app/config.yaml` atau env variable
- Expose port sesuai `webhook_port`

---

## Error Handling & Logging

- Gunakan `log/slog` (standar Go 1.21+) untuk structured logging
- Level: INFO untuk aksi yang diambil (pesan dihapus, user di-ban), ERROR untuk kegagalan API
- Jangan log konten pesan pengguna (privasi)
- Semua error dari Telegram API di-log tapi tidak menyebabkan panic

---

## README.md (minimal)

Wajib berisi:
- Cara set up webhook di Telegram (`setWebhook` API call)
- Cara konfigurasi via `config.yaml`
- Cara konfigurasi via environment variable
- Cara jalankan dengan Docker
- Cara jalankan langsung (`go run`)
- Daftar lengkap semua opsi konfigurasi beserta penjelasan dan nilai defaultnya