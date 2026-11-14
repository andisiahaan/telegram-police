package admin

import (
	"html/template"

	"github.com/andisiahaan/telegram-police/internal/store"
)

type DashboardData struct {
	ForbiddenWords      []store.WordItem
	ForbiddenTitleWords []store.TitleWordItem
	WhitelistSenders    []store.WhitelistItem
	TotalWords          int
	TotalTitleWords     int
	TotalWhitelist      int
	FlashMessage        string
	FlashError          string
}

const loginHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>TelePolice Admin - Login</title>
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
      background: #0f172a; color: #f8fafc; min-height: 100vh;
      display: flex; align-items: center; justify-content: center; padding: 20px;
    }
    .card {
      background: #1e293b; border: 1px solid #334155; border-radius: 12px;
      padding: 32px; width: 100%; max-width: 380px; box-shadow: 0 10px 25px rgba(0,0,0,0.4);
    }
    h1 { font-size: 1.5rem; font-weight: 700; margin-bottom: 8px; text-align: center; }
    p.sub { font-size: 0.875rem; color: #94a3b8; text-align: center; margin-bottom: 24px; }
    .alert {
      background: #ef444420; border: 1px solid #ef444480; color: #fca5a5;
      padding: 10px 14px; border-radius: 8px; font-size: 0.875rem; margin-bottom: 16px;
    }
    label { display: block; font-size: 0.875rem; font-weight: 500; margin-bottom: 6px; color: #cbd5e1; }
    input[type="password"] {
      width: 100%; padding: 12px; background: #0f172a; border: 1px solid #475569;
      border-radius: 8px; color: #fff; font-size: 1.1rem; text-align: center;
      letter-spacing: 4px; margin-bottom: 20px; outline: none;
    }
    input[type="password"]:focus { border-color: #38bdf8; ring: 2px #38bdf8; }
    button {
      width: 100%; padding: 12px; background: #2563eb; color: #fff;
      border: none; border-radius: 8px; font-size: 1rem; font-weight: 600;
      cursor: pointer; transition: background 0.2s;
    }
    button:hover { background: #1d4ed8; }
  </style>
</head>
<body>
  <div class="card">
    <h1>👮‍♂️ TelePolice</h1>
    <p class="sub">Enter Admin PIN to access dashboard</p>
    {{if .Error}}
      <div class="alert">{{.Error}}</div>
    {{end}}
    <form method="POST" action="/admin/login">
      <label for="pin">Security PIN</label>
      <input type="password" id="pin" name="pin" autofocus required placeholder="••••••">
      <button type="submit">Unlock Dashboard</button>
    </form>
  </div>
</body>
</html>`

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>TelePolice Dashboard</title>
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
      background: #090d16; color: #f1f5f9; min-height: 100vh; padding: 24px 16px;
    }
    .container { max-width: 1000px; margin: 0 auto; }
    header {
      display: flex; justify-content: space-between; align-items: center;
      margin-bottom: 28px; padding-bottom: 16px; border-bottom: 1px solid #1e293b;
    }
    .title-area h1 { font-size: 1.5rem; font-weight: 700; display: flex; align-items: center; gap: 8px; }
    .status-badge {
      display: inline-block; font-size: 0.75rem; background: #059669; color: #ecfdf5;
      padding: 3px 8px; border-radius: 9999px; font-weight: 600; margin-left: 8px;
    }
    .btn-logout {
      background: #334155; color: #cbd5e1; border: none; padding: 8px 14px;
      border-radius: 6px; cursor: pointer; font-size: 0.875rem; font-weight: 500;
    }
    .btn-logout:hover { background: #475569; }
    .stats-grid {
      display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 16px; margin-bottom: 28px;
    }
    .stat-card {
      background: #111827; border: 1px solid #1f2937; border-radius: 10px; padding: 18px;
    }
    .stat-label { font-size: 0.8rem; color: #9ca3af; text-transform: uppercase; font-weight: 600; margin-bottom: 4px; }
    .stat-value { font-size: 1.8rem; font-weight: 800; color: #38bdf8; }
    .card {
      background: #111827; border: 1px solid #1f2937; border-radius: 12px;
      padding: 24px; margin-bottom: 24px; box-shadow: 0 4px 6px -1px rgba(0,0,0,0.3);
    }
    .card h2 { font-size: 1.15rem; font-weight: 700; margin-bottom: 6px; color: #f8fafc; }
    .card p.desc { font-size: 0.85rem; color: #94a3b8; margin-bottom: 16px; }
    .form-row { display: flex; gap: 10px; margin-bottom: 20px; }
    .form-row input[type="text"], .form-row input[type="number"] {
      flex: 1; padding: 10px 14px; background: #030712; border: 1px solid #374151;
      border-radius: 6px; color: #fff; font-size: 0.95rem; outline: none;
    }
    .form-row input:focus { border-color: #3b82f6; }
    .btn-add {
      background: #2563eb; color: #fff; border: none; padding: 10px 18px;
      border-radius: 6px; font-weight: 600; cursor: pointer; white-space: nowrap;
    }
    .btn-add:hover { background: #1d4ed8; }
    .tag-container { display: flex; flex-wrap: wrap; gap: 8px; max-height: 260px; overflow-y: auto; padding: 4px; }
    .tag {
      background: #1f2937; border: 1px solid #374151; padding: 6px 12px;
      border-radius: 6px; display: inline-flex; align-items: center; gap: 8px; font-size: 0.9rem;
    }
    .tag span { font-family: monospace; }
    .btn-del {
      background: none; border: none; color: #ef4444; cursor: pointer;
      font-size: 1rem; line-height: 1; font-weight: bold;
    }
    .btn-del:hover { color: #f87171; }
    table { width: 100%; border-collapse: collapse; font-size: 0.9rem; }
    th { text-align: left; padding: 10px; background: #1f2937; color: #94a3b8; font-weight: 600; }
    td { padding: 10px; border-bottom: 1px solid #1f2937; color: #e2e8f0; }
    .empty-state { color: #64748b; font-size: 0.875rem; font-style: italic; }
  </style>
</head>
<body>
  <div class="container">
    <header>
      <div class="title-area">
        <h1>👮‍♂️ TelePolice Admin <span class="status-badge">Active</span></h1>
      </div>
      <form method="POST" action="/admin/logout">
        <button type="submit" class="btn-logout">Logout</button>
      </form>
    </header>

    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-label">Kata Pesan Terlarang</div>
        <div class="stat-value">{{.TotalWords}}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Kata Username Terlarang</div>
        <div class="stat-value">{{.TotalTitleWords}}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Whitelist Pengirim</div>
        <div class="stat-value">{{.TotalWhitelist}}</div>
      </div>
    </div>

    <!-- 1. Forbidden Words in Message -->
    <div class="card">
      <h2>Forbidden Words (Isi Pesan)</h2>
      <p class="desc">Pesan di grup yang memuat kata-kata ini akan otomatis dihapus dan dikenai tindakan moderasi. Pisahkan dengan koma untuk menambah banyak kata sekaligus.</p>
      <form class="form-row" method="POST" action="/admin/words">
        <input type="text" name="words" placeholder="Contoh: judi, slot, bokep" required>
        <button type="submit" class="btn-add">Tambah Kata</button>
      </form>
      <div class="tag-container">
        {{if .ForbiddenWords}}
          {{range .ForbiddenWords}}
            <div class="tag">
              <span>{{.Word}}</span>
              <form method="POST" action="/admin/words/delete" style="display:inline;">
                <input type="hidden" name="id" value="{{.ID}}">
                <button type="submit" class="btn-del" title="Hapus kata">&times;</button>
              </form>
            </div>
          {{end}}
        {{else}}
          <p class="empty-state">Belum ada kata terlarang.</p>
        {{end}}
      </div>
    </div>

    <!-- 2. Forbidden Title Words -->
    <div class="card">
      <h2>Forbidden Title Words (Nama Pengguna / Username)</h2>
      <p class="desc">Pesan dari pengguna yang nama atau username-nya memuat kata-kata ini akan diblokir.</p>
      <form class="form-row" method="POST" action="/admin/title-words">
        <input type="text" name="words" placeholder="Contoh: admin, moderator, support" required>
        <button type="submit" class="btn-add">Tambah Kata Title</button>
      </form>
      <div class="tag-container">
        {{if .ForbiddenTitleWords}}
          {{range .ForbiddenTitleWords}}
            <div class="tag">
              <span>{{.Word}}</span>
              <form method="POST" action="/admin/title-words/delete" style="display:inline;">
                <input type="hidden" name="id" value="{{.ID}}">
                <button type="submit" class="btn-del" title="Hapus kata">&times;</button>
              </form>
            </div>
          {{end}}
        {{else}}
          <p class="empty-state">Belum ada kata title terlarang.</p>
        {{end}}
      </div>
    </div>

    <!-- 3. Whitelist Senders -->
    <div class="card">
      <h2>Whitelist Senders (Pengecualian ID Pengguna)</h2>
      <p class="desc">Pengguna dengan ID Telegram di bawah ini akan dibypass dari seluruh filter moderasi.</p>
      <form class="form-row" method="POST" action="/admin/whitelist">
        <input type="number" name="user_id" placeholder="Telegram User ID (e.g. 123456789)" required>
        <input type="text" name="description" placeholder="Keterangan (opsional, misal: Bot Notifikasi)">
        <button type="submit" class="btn-add">Tambah Whitelist</button>
      </form>
      {{if .WhitelistSenders}}
        <table>
          <thead>
            <tr>
              <th>User ID</th>
              <th>Keterangan</th>
              <th>Aksi</th>
            </tr>
          </thead>
          <tbody>
            {{range .WhitelistSenders}}
              <tr>
                <td><code>{{.UserID}}</code></td>
                <td>{{.Description}}</td>
                <td>
                  <form method="POST" action="/admin/whitelist/delete" style="display:inline;">
                    <input type="hidden" name="id" value="{{.ID}}">
                    <button type="submit" class="btn-del" title="Hapus whitelist">&times;</button>
                  </form>
                </td>
              </tr>
            {{end}}
          </tbody>
        </table>
      {{else}}
        <p class="empty-state">Belum ada user ID di whitelist.</p>
      {{end}}
    </div>
  </div>
</body>
</html>`

var (
	loginTmpl     = template.Must(template.New("login").Parse(loginHTML))
	dashboardTmpl = template.Must(template.New("dashboard").Parse(dashboardHTML))
)
