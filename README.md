<div align="center">

# 🤖 Gowa-Bot

**WhatsApp Bot sederhana dan powerful yang dibangun dengan Go**

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg?style=for-the-badge)](LICENSE)
[![WhatsApp](https://img.shields.io/badge/WhatsApp-25D366?style=for-the-badge&logo=whatsapp&logoColor=white)](https://whatsapp.com/)

![Banner](https://files.catbox.moe/1xnz38.jpg)

[Fitur](#-fitur) • [Instalasi](#-instalasi) • [Penggunaan](#-penggunaan) • [Command](#-daftar-command) • [Konfigurasi](#-konfigurasi)

</div>

---

## 📖 Tentang Proyek

**Gowa-Bot** adalah bot WhatsApp yang dibuat dengan ❤️ menggunakan library [Gowa](https://github.com/jrevanaldi-ai/gowa). Bot ini dirancang untuk menjadi ringan, cepat, dan mudah dikustomisasi sesuai kebutuhan Anda.

Dengan arsitektur yang modular, Anda dapat dengan mudah menambahkan command baru atau memodifikasi fitur yang sudah ada. Cocok untuk kebutuhan personal maupun grup WhatsApp Anda!

---

## ✨ Fitur

| 🚀 | **Ringan & Cepat** | Dibangun dengan Go, bot ini sangat efisien dalam penggunaan resource |
|----|-------------------|---------------------------------------------------------------------|
| 📦 | **Modular** | Sistem command yang terstruktur, mudah untuk menambah fitur baru |
| 🔐 | **Owner System** | Kontrol akses berbasis owner untuk command sensitif |
| 💬 | **Reply Message** | Support reply pesan dengan format yang rapih |
| 🎭 | **Ephemeral Message** | Pesan hilang otomatis di group yang mendukung |
| 🗄️ | **SQLite Database** | Session dan data app disimpan lokal, aman dan persisten |
| 🎨 | **Formatted Output** | Pesan dengan format menarik dan mudah dibaca |
| ⚡ | **Multi-threading** | Handle multiple pesan secara concurrent |
| 🤝 | **Jadibot** | Multi-bot — user lain bisa pairing nomor mereka sebagai sub-bot |
| 🧠 | **AI Integration** | Integrasi Claude AI lewat keyword `lune` |
| 💰 | **Payment Gateway** | Donasi QRIS via MustikaPay |
| 🎵 | **Music Player** | Play audio dari YouTube & Spotify via API yardansh |
| ⬇️ | **Downloader** | Download dari Instagram, TikTok, GitHub |
| 🎨 | **Sticker Maker** | Brat (dari teks) & Image→Sticker (kirim/reply gambar dengan `.s`) — auto WebP 512×512 |
| 🚫 | **Ban System** | Ban user atau group dari pemakaian bot |
| 🛡️ | **Eval Sandbox** | Eksekusi kode Go runtime via yaegi (owner only) |
| 📊 | **Web Dashboard** | Live dashboard di `:8080` — push realtime via SSE, no polling |

---

## 📋 Prasyarat

Sebelum memulai, pastikan Anda telah menginstal:

- **[Go](https://go.dev/dl/)** versi 1.26 atau lebih tinggi
- **[Git](https://git-scm.com/downloads)** untuk clone repository
- **WhatsApp** aktif untuk pairing bot
- **Image converter** (salah satu saja) — wajib untuk fitur sticker (`.brat`, `.s`). Helper otomatis pilih binary yang tersedia: `magick` (ImageMagick 7) → `convert` (ImageMagick 6) → `ffmpeg` (dengan `libwebp`).
  - Debian/Ubuntu: `sudo apt install imagemagick` **atau** `sudo apt install ffmpeg`

---

## 🚀 Instalasi

### 1. Clone Repository

```bash
git clone https://github.com/jrevanaldi-ai/gowa-bot.git
cd gowa-bot
```

### 2. Download Dependencies

```bash
go mod download
```

> 💡 **Catatan:** Project ini menggunakan fork lokal dari library Gowa di folder `gowa-lib/`. Direktori ini sudah disertakan, tidak perlu clone terpisah.

### 3. Setup Environment

Salin file `.env.example` menjadi `.env`:

```bash
cp .env.example .env
```

Kemudian edit file `.env` dan sesuaikan konfigurasi:

```bash
# Owner numbers (gunakan format internasional tanpa +)
export GOWA_BOT_OWNERS="6281234567890"

# Database path (opsional)
export GOWA_BOT_DB="gowa-bot.db"

# Log level (debug, info, warn, error)
export GOWA_BOT_LOG_LEVEL="info"

# Self mode (true/false)
export GOWA_BOT_SELF_MODE="false"

# MustikaPay API key untuk fitur donasi (opsional)
export GOWA_BOT_MUSTIKA_API_KEY=""

# AI API key untuk fitur .lune / keyword lune (opsional)
export GOWA_BOT_AI_API_KEY=""
```

### 4. Build & Run

```bash
# Build aplikasi
go build -o gowa-bot

# Jalankan bot
./gowa-bot -phone 6281234567890
```

> 💡 **Tips:** Gunakan flag `-phone` dengan nomor WhatsApp Anda (format internasional tanpa tanda `+`)

---

## 📖 Penggunaan

### First Time Setup

Saat pertama kali menjalankan bot, Anda perlu melakukan **pairing**:

1. Jalankan bot dengan flag `-phone`:
   ```bash
   ./gowa-bot -phone 6281234567890
   ```

2. Bot akan menampilkan **pairing code** (8 karakter)

3. Buka WhatsApp di HP Anda → **Perangkat Tertaut** → **Tautkan Perangkat**

4. Masukkan pairing code yang ditampilkan di terminal

5. ✅ Selesai! Bot siap digunakan

### Command Flags

| Flag | Deskripsi | Contoh |
|------|-----------|--------|
| `-phone` | Nomor WhatsApp untuk pairing | `-phone 6281234567890` |
| `-pair` | Pairing code custom (opsional) | `-pair ABCD1234` |
| `-db` | Path database SQLite | `-db ./data/bot.db` |
| `-log-level` | Level logging | `-log-level debug` |
| `-self` | Self mode - bot merespon pesan sendiri | `-self` |
| `-mustika-api-key` | API key MustikaPay (override env) | `-mustika-api-key xxx` |
| `-ai-api-key` | API key Claude AI (override env) | `-ai-api-key sk-ant-xxx` |
| `-web-addr` | Bind address dashboard HTTP | `-web-addr :8080` atau `127.0.0.1:8080` |

> 💡 Semua flag punya fallback ke environment variable dengan prefix `GOWA_BOT_`.

---

## 📜 Daftar Command

Bot menggunakan **prefix** `.` untuk command (bisa diganti dengan `.setprefix`). Owner dapat menggunakan command **tanpa prefix**.

> 💡 **Kategori (tag) di `.menu`** mengikuti field `Tag` pada metadata command. Tag yang sekarang dipakai: `main`, `utility`, `play`, `download`, `maker`, `search`, `owner`, `jadibot`, `debug`. `.menu` menampilkan tag-nya secara UPPER-CASE dan sorted alfabet.

### 🛠️ Utility

| Command | Alias | Deskripsi | Contoh |
|---------|-------|-----------|--------|
| `.ping` | `.p` | Cek latency bot | `.ping` |
| `.lune` | - | Tanya AI Claude | `.lune apa itu Go?` |
| `.lune reset` | - | Reset history percakapan AI | `.lune reset` |

### 📋 Main

| Command | Alias | Deskripsi | Contoh |
|---------|-------|-----------|--------|
| `.menu` | `.m`, `.h` | Tampilkan daftar command | `.menu` |
| `.help` | `.info`, `.command` | Lihat detail command | `.help ping` |
| `.getpp` | `.pp` | Ambil foto profil user | `.getpp @user` |
| `.donasi` | `.donate` | Buat QRIS donasi (butuh MustikaPay) | `.donasi 10000` |
| `.cekdonasi` | `.checkdonate` | Cek status donasi | `.cekdonasi <refno>` |

> 💬 **Trigger AI tanpa prefix:**
> - **Keyword `lune`:** pesan apa pun yang mengandung kata "lune" (case-insensitive, tanpa prefix sekalipun) otomatis dijawab AI. Contoh: `"hai lune apa kabar?"`
> - **Reply ke pesan AI:** balas pesan dari bot AI (dalam 30 menit terakhir, tracked via cache) juga otomatis dilanjutkan sebagai percakapan AI tanpa perlu nulis `lune` lagi.

### 🎵 Play

| Command | Alias | Deskripsi | Contoh |
|---------|-------|-----------|--------|
| `.play` | `.ytmp3`, `.yta` | Cari & kirim audio dari YouTube (API yardansh) | `.play Multo Cup of Joe` |
| `.spotify` | `.sp`, `.splay` | Cari & kirim audio dari Spotify (API yardansh) | `.spotify Multo` |

### ⬇️ Download

| Command | Alias | Deskripsi | Contoh |
|---------|-------|-----------|--------|
| `.instagram` | - | Download Instagram post/reel | `.instagram <url>` |
| `.tiktok` | - | Download video TikTok | `.tiktok <url>` |
| `.github` | - | Info / download repo GitHub | `.github user/repo` |

### 🎨 Maker

| Command | Alias | Deskripsi | Contoh |
|---------|-------|-----------|--------|
| `.brat` | - | Buat sticker brat dari teks (PNG → WebP 512×512) | `.brat halo dunia` |
| `.sticker` | `.s`, `.stiker` | Convert gambar jadi sticker (caption / reply) | kirim gambar dengan caption `.s` atau reply gambar dengan `.s` |

> 💡 Butuh image converter terpasang (`magick` / `convert` / `ffmpeg`). Helper di `helper/sticker.go` otomatis pilih yang tersedia — shared antara `.brat` dan `.s`.

### 🔍 Search

| Command | Alias | Deskripsi | Contoh |
|---------|-------|-----------|--------|
| `.ttsearch` | - | Search video TikTok | `.ttsearch keyword` |

### 🔐 Owner Only

| Command | Alias | Deskripsi | Contoh |
|---------|-------|-----------|--------|
| `$<cmd>` | - | Eksekusi shell command (reserved prefix) | `$ls -la` |
| `>>` / `.eval` | `ev` | Eksekusi kode Go runtime (yaegi) | `>> return ctx.Chat.String()` |
| `.setmode` | - | Ganti mode bot (self/public) | `.setmode self` |
| `.setprefix` | - | Ganti prefix command | `.setprefix !` |
| `.infoserver` | - | Info server (CPU/RAM/uptime) | `.infoserver` |
| `.react` | - | Reaksi emoji ke pesan | `.react ❤️` |
| `.join` | `.joingrup`, `.joingroup` | Join grup via invite link | `.join https://chat.whatsapp.com/ABC123` |
| `.bangroup` | - | Ban group dari pemakaian bot | `.bangroup` |
| `.unbangroup` | - | Unban group | `.unbangroup` |
| `.banuser` | - | Ban user | `.banuser @user` |
| `.unbanuser` | - | Unban user | `.unbanuser @user` |

> 💡 Owner dapat memanggil command tanpa prefix `.` (kecuali pesan `IsFromMe` — anti-loop). Prefix `$` di-reserve khusus untuk `exec` dan tidak bisa diganti via `.setprefix`.

### 🤝 Jadibot (Multi-Bot)

Memungkinkan user lain pairing nomor mereka sebagai sub-bot di bawah Gowa-Bot utama.

| Command | Alias | Deskripsi | Contoh |
|---------|-------|-----------|--------|
| `.jadibot` | - | Daftar nomor jadi sub-bot | `.jadibot 6281234567890` |
| `.listjadibot` | - | List semua jadibot aktif | `.listjadibot` |
| `.stopjadibot` | - | Hentikan jadibot | `.stopjadibot <id>` |
| `.pausejadibot` | - | Pause jadibot (bisa di-resume) | `.pausejadibot <id>` |
| `.resumejadibot` | - | Resume jadibot yang di-pause | `.resumejadibot <id>` |
| `.deletejadibot` | `.deljb`, `.deletejb`, `.delsession` | **User biasa** hapus jadibot **milik sendiri** (verifikasi via `OwnerJID.User`) | `.deletejadibot` (auto kalau hanya 1) atau `.deletejadibot <id>` |
| `.removejadibot` | - | **Owner bot induk** hapus jadibot siapapun (permanen) | `.removejadibot <id>` |

> 💡 Beda `.deletejadibot` vs `.removejadibot`:
> - `.deletejadibot` → tag `owner` di registry, tapi `OwnerOnly: false`. Verifikasi ownership di handler. Cocok untuk **user kreator** hapus session-nya sendiri.
> - `.removejadibot` → `OwnerOnly: true`. Hanya owner bot induk, bisa hapus jadibot user manapun (paksa hapus).

### 🐞 Debug

| Command | Alias | Deskripsi | Contoh |
|---------|-------|-----------|--------|
| `.checkephemeral` | `.ce` | Cek status ephemeral group | `.checkephemeral` |

---

## 📊 Web Dashboard

Dashboard live di `http://localhost:8080` (default; override via flag `-web-addr`). **Realtime via Server-Sent Events** — server push update setiap detik, browser tinggal subscribe `EventSource`. Tidak ada polling.

### Endpoints

| Endpoint | Tipe | Deskripsi |
|---|---|---|
| `/` | static | Embedded dashboard SPA (HTML/CSS/JS via `//go:embed web`) |
| `/api/info` | JSON | Hostname, uptime, Go version, OS/arch, CPU, goroutines |
| `/api/memory` | JSON | RSS, heap, stack, GC count, last GC |
| `/api/bot` | JSON | Status koneksi WA, phone/JID, push name, self mode, prefixes |
| `/api/commands` | JSON | Semua command registered (cmd, tag, desc, alias, owner_only) |
| `/api/jadibots` | JSON | Daftar jadibot aktif dari DB |
| `/api/stream` | **SSE** | Push event: `init` (semua), `tick` (dynamic tiap 1s), `commands` (tiap 10s) |

### Yang ditampilkan dashboard

- **Server card** — hostname, uptime, started, platform, Go version, CPU, goroutines
- **Memory card** — RSS, heap in-use/idle, stack, sys total, total alloc (cumulative), GC count, last GC (relative)
- **Bot card** — status (Online / Disconnected / Not paired), phone, **full JID**, push name, self mode, prefixes
- **Jadibot table** — ID (shortened), phone, owner, status, running
- **Commands** — di-group by tag, hover untuk lihat desc+example, command owner-only diberi warna merah

UI fully responsive (320px → 1280px+), dark text di canvas terang, support touch device.

### Mode Bot

Bot ini memiliki 2 mode operasi:

| Mode | Deskripsi | Cara Aktivasi |
|------|-----------|---------------|
| **Public Mode** (Default) | Bot hanya merespon pesan dari user lain | `.setmode public` |
| **Self Mode** | Bot merespon pesan dari diri sendiri **hanya jika nomor bot terdaftar sebagai owner** | `.setmode self` |

> 💡 **Tips:** Owner bisa mengganti mode kapan saja langsung dari WhatsApp tanpa perlu restart bot!
>
> 🔒 **Keamanan:** Self mode dirancang untuk testing/development. Hanya owner yang bisa menggunakan command `.setmode`. Jika nomor bot tidak terdaftar sebagai owner, self mode tidak akan berfungsi.

### Contoh Penggunaan

<details>
<summary><b>📋 Menu Command</b></summary>

```
Kirim: .menu

Output:
GOWA-BOT

DEBUG:
- checkephemeral (ce)

DOWNLOAD:
- github
- instagram
- tiktok

JADIBOT:
- jadibot
- listjadibot
- pausejadibot
- removejadibot
- resumejadibot
- stopjadibot

MAIN:
- cekdonasi (checkdonate)
- donasi (donate)
- getpp (pp)
- help (info)

MAKER:
- brat
- sticker (s)

OWNER:
- bangroup
- banuser
- deletejadibot (deljadibot)
- eval (ev)
- exec
- infoserver
- join (joingrup)
- react
- setmode
- setprefix
- unbangroup
- unbanuser

PLAY:
- play (ytmp3)
- spotify (sp)

SEARCH:
- ttsearch

UTILITY:
- lune
- ping (p)
```

> Tag di-sort alfabetis dan command di tiap tag juga sorted. Hanya command pertama dari tiap meta yang ditampilkan dengan alias pertama dalam tanda kurung.

</details>

<details>
<summary><b>🏓 Ping Command</b></summary>

```
Kirim: .ping

Output (dikirim pertama):
Pong

Latency: calculating...
Status: Online
Uptime: 00:15:32

Lalu di-edit otomatis jadi:
Pong

Latency: 45 ms
Status: Online
Uptime: 00:15:32
```

</details>

<details>
<summary><b>🧠 AI Lune</b></summary>

```
Trigger 1 — pakai command:
Kirim: .lune jelaskan apa itu goroutine

Trigger 2 — keyword:
Kirim: hey lune, apa kabar?

Trigger 3 — reply ke pesan AI sebelumnya (auto-continue):
Reply pesan bot AI dengan teks apa saja → otomatis dilanjut sebagai percakapan AI

Reset history percakapan:
Kirim: .lune reset

Output:
[Respon AI Claude]
```

> 🔒 History percakapan disimpan per-user di in-memory cache dengan TTL **30 menit**, max **15 pesan terakhir**. Tracking reply juga TTL 30 menit dan disimpan per chat JID.

</details>

<details>
<summary><b>🎨 Sticker dari Gambar</b></summary>

```
Cara 1 — caption:
Kirim gambar dengan caption: .s
→ Bot balas dengan sticker 512×512

Cara 2 — reply:
Reply ke pesan gambar dengan teks: .s
→ Bot balas dengan sticker
```

> Author di EXIF sticker pack otomatis pakai `PushName` pengirim. Cocok juga untuk gambar yang sudah lewat di chat — tinggal reply dengan `.s`.

</details>

<details>
<summary><b>🤝 Jadibot</b></summary>

```
Kirim: .jadibot 6289876543210

Output:
✓ Pairing code: ABCD-1234
Masukkan kode di WhatsApp 6289876543210
→ Perangkat Tertaut → Tautkan Perangkat
```

Setelah pairing berhasil, nomor 6289876543210 akan jadi sub-bot dengan command yang sama. Session disimpan di `sessions/jadibot_<uuid>/`.

</details>

<details>
<summary><b>💻 Exec Command (Owner Only)</b></summary>

```
Kirim: $ls -la

Output:
✓ Output:
total 48
drwxr-xr-x  8 user user 4096 Mar 23 10:00 .
-rw-r--r--  1 user user  234 Mar 23 10:00 .env
...
```

</details>

<details>
<summary><b>🛡️ Eval Command (Owner Only)</b></summary>

```
Kirim: >> return fmt.Sprintf("Chat: %s", ctx.Chat)

Output:
✓ Result: Chat: 120363xxx@g.us
```

Variable pre-bound: `ctx` (CommandContext), `c` (gowa.Client), `db` (DatabaseManager).

</details>

<details>
<summary><b>ℹ️ Help Command</b></summary>

```
Kirim: .help ping

Output:
HELP: PING

Category: Utility
Description: Cek respon bot dan latency
Command: .ping
Aliases: .p
Example: .ping
Access: Public
```

> `.help` tanpa argumen akan fallback ke `.menu`.

</details>

---

## ⚙️ Konfigurasi

### Environment Variables

| Variable | Deskripsi | Default | Required |
|----------|-----------|---------|----------|
| `GOWA_BOT_OWNERS` | Daftar nomor owner (comma separated) | - | ✅ Ya |
| `GOWA_BOT_DB` | Path database SQLite | `gowa-bot.db` | ❌ Tidak |
| `GOWA_BOT_LOG_LEVEL` | Level logging (debug/info/warn/error) | `info` | ❌ Tidak |
| `GOWA_BOT_SELF_MODE` | Self mode (true/false) | `false` | ❌ Tidak |
| `GOWA_BOT_MUSTIKA_API_KEY` | API key MustikaPay (donasi QRIS) | - | ❌ Tidak |
| `GOWA_BOT_AI_API_KEY` | API key Claude AI (`.lune`) | - | ❌ Tidak |

> 💡 Fitur yang butuh API key (MustikaPay, AI) hanya aktif jika key tersedia. Bot tetap jalan normal kalau dikosongkan.

### Format Nomor Owner

Gunakan format internasional **tanpa** tanda `+` atau spasi:

```bash
# ✅ Benar
export GOWA_BOT_OWNERS="6281234567890"
export GOWA_BOT_OWNERS="6281234567890,6289876543210"

# ❌ Salah
export GOWA_BOT_OWNERS="+62 812-3456-7890"
export GOWA_BOT_OWNERS="081234567890"
```

---

## 🏗️ Struktur Proyek

```
gowa-bot/
├── 📄 main.go              # Entry point — bootstrap, flag, koneksi WA
├── 📄 registryCmd.go       # registerCommands() — daftarkan command di sini
├── 📦 go.mod               # Dependensi Go (replace gowa → ./gowa-lib)
├── 🔧 .env.example         # Template konfigurasi
├── 📖 README.md            # Dokumentasi (file ini)
├── 📖 CLAUDE.md            # Guide untuk Claude Code
│
├── 📂 client/
│   └── bot_client.go         # Event handler & message dispatcher
│
├── 📂 commands/
│   ├── 📂 general/           # menu, help, getpp, donasi, lune (AI)
│   ├── 📂 utility/           # ping
│   ├── 📂 owner/             # exec, eval, setmode, setprefix, ban, deletejadibot (self-service)
│   ├── 📂 jadibot/           # multi-bot management (jadibot, list/stop/pause/resume/remove)
│   ├── 📂 download/          # play, spotify (tag=play), instagram, tiktok, github, ttsearch
│   ├── 📂 maker/             # brat (text→sticker), sticker (image→sticker)
│   └── 📂 debug/             # checkephemeral
│
├── 📂 helper/
│   ├── logger.go             # Logger berwarna (Message multiline, Debug silent)
│   ├── cache.go              # Cache TTL in-memory
│   ├── ephemeral.go          # Helper pesan ephemeral
│   ├── message.go            # Builder reply message (CreateSimpleReply)
│   ├── database.go           # SQLite manager (jadibot/banned/donasi)
│   ├── session_manager.go    # Jadibot session manager
│   ├── sticker.go            # Convert image → WebP 512×512 + EXIF metadata + upload (shared brat & .s)
│   ├── ai.go                 # Claude AI service
│   ├── ai_reply.go           # Tracking message-ID balasan AI (cache 30 menit)
│   ├── url.go                # URL utilities (ExtractWhatsAppInviteCode, dll)
│   └── mustikapay.go         # MustikaPay QRIS integration
│
├── 📂 lib/
│   ├── types.go              # CommandRegistry, CommandContext, interfaces
│   └── dispatcher.go         # Semaphore worker pool (panic-recover) — dipakai BotClient.HandleMessage
│
├── 📂 webserver/
│   ├── server.go             # HTTP server + SSE stream endpoint (`/api/stream`)
│   └── 📂 web/               # Embedded dashboard SPA (HTML/CSS/JS + logo.png)
│
├── 📂 gowa-lib/              # Fork lokal library Gowa
└── 📂 sessions/              # Storage untuk session jadibot (auto-generated)
```

> 💡 File command terorganisir berdasarkan **folder**, sedangkan kategori di `.menu` terorganisir berdasarkan **field `Tag`** di metadata. `play.go` & `spotify.go` ada di folder `download/` tapi Tag-nya `"play"`, jadi muncul di section `PLAY:` saat `.menu`.

---

## 🛠️ Development

### Menambah Command Baru

1. Buat file baru di folder `commands/` sesuai kategori:

```go
// commands/utility/halo.go
package utility

import (
    "github.com/jrevanaldi-ai/gowa-bot/helper"
    "github.com/jrevanaldi-ai/gowa-bot/lib"
)

var HaloMetadata = &lib.CommandMetadata{
    Cmd:       "halo",
    Tag:       "utility",
    Desc:      "Ucapkan salam",
    Example:   ".halo",
    Hidden:    false,
    OwnerOnly: false,
    Alias:     []string{"hi"},
}

func HaloHandler(ctx *lib.CommandContext) error {
    message := "Halo! Apa kabar? 👋"
    _, err := ctx.SendMessage(helper.CreateSimpleReply(
        message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String(),
    ))
    return err
}
```

2. **Wajib** daftarkan di `registryCmd.go::registerCommands()`:

```go
func registerCommands(registry *lib.CommandRegistry) {
    // ... existing commands
    registry.Register(utility.HaloMetadata, utility.HaloHandler)
}
```

> ⚠️ **Tidak ada autoload.** Kalau lupa daftar di `registryCmd.go`, command tidak akan jalan. Import package command-nya juga harus ditambahkan di atas file `registryCmd.go`.

3. Build dan jalankan!

### Arsitektur Singkat

Alur pesan masuk:

```
WhatsApp Event
    ↓
BotClient.EventHandler (client/bot_client.go)
    ↓
HandleMessage (goroutine) → processMessage()
    ├─ Cek self-mode & IsFromMe
    ├─ Ekstrak teks dari semua tipe pesan (text/image caption/video caption/dll)
    ├─ Unwrap edited / ephemeral / view-once / document-with-caption (recursive)
    ├─ Cek ban via DBManager.IsBanned (skip kalau owner) — cek group JID & user JID
    ├─ Special: $<cmd> oleh owner → handleExecCommand (langsung ke handler `exec`)
    ├─ Extract reply context (StanzaID + Participant + QuotedMessage text) + mentions
    ├─ parseCommandWithOwner — prefix list / owner-tanpa-prefix (skip kalau IsFromMe)
    ├─ Fallback: kalau bukan command tapi mengandung "lune" atau reply ke pesan AI → handler `lune`
    ├─ Validasi OwnerOnly
    └─ Jalankan handler dengan CommandContext (Client, BotClient, Sender, Chat, Args, ReplyMessage, Mentions, EphemeralWrapper)
```

Untuk detail teknis lebih lanjut, lihat **[CLAUDE.md](CLAUDE.md)**.

### Tips Development

- **Toolchain Go:** Proyek ini di-pin ke Go 1.26 (`go.mod`, README, CI). Pastikan toolchain lokal Anda match sebelum nuduh bug kode kalau `go build` gagal.
- **Fork Gowa:** Direktori `gowa-lib/` adalah fork lokal (lihat `replace` di `go.mod`). Edit di sana langsung mempengaruhi bot — `go mod tidy` tidak akan pull versi remote.
- **Session DB:** Jangan hapus `gowa-bot.db` saat bot jalan, session WhatsApp **dan** data app (jadibots/banned/donations) ada di file yang sama.
- **Cyclic import:** `client/` import `commands/owner` (untuk `ParseExecCommand`). Jangan bikin `commands/owner` import `client/`. Pakai `lib.BotClientInterface` / `lib.JadibotSessionManagerInterface` kalau butuh akses.
- **Jadibot client factory:** Saat bikin variasi `BotClient` untuk jadibot, gunakan `ClientFactory` closure di `main.go` — pola ini mencegah cycle antara `helper/` ↔ `client/`.
- **Owner matching di jadibot:** Child jadibot dibuat dengan list owner kosong; deteksi owner di child hanya mengandalkan match `Client.Store.ID.User` (alias: "owner = nomor jadibot itu sendiri").

---

## 🤝 Kontribusi

Kontribusi sangat diapresiasi! Berikut cara berkontribusi:

1. **Fork** repository ini
2. Buat **Feature Branch** (`git checkout -b feature/AmazingFeature`)
3. **Commit** perubahan (`git commit -m 'Add some AmazingFeature'`)
4. **Push** ke branch (`git push origin feature/AmazingFeature`)
5. Buka **Pull Request**

### Guidelines

- Ikuti style code yang sudah ada
- Tambahkan komentar untuk logic yang kompleks
- Test command baru sebelum submit PR
- Update dokumentasi jika diperlukan

---

## 📄 License

Proyek ini dilisensikan di bawah **MIT License**. Lihat file [LICENSE](LICENSE) untuk detail lebih lanjut.

---

## 🙏 Ucapan Terima Kasih

Terima kasih kepada:

- **[Gowa Library](https://github.com/jrevanaldi-ai/gowa)** - WhatsApp client library yang powerful
- **[yaegi](https://github.com/traefik/yaegi)** - Go interpreter untuk fitur eval
- **[Claude AI](https://www.anthropic.com/)** - AI engine di balik `.lune`
- **[Go Community](https://go.dev/)** - Komunitas Go yang luar biasa
- **Semua contributor** yang telah berkontribusi dalam pengembangan bot ini

---

## 📞 Support

Jika Anda mengalami masalah atau memiliki pertanyaan:

- 📧 Buka **Issue** di repository ini
- 💬 Diskusi di **Discussions** tab
- 📖 Cek dokumentasi Gowa Library

---

<div align="center">

**Dibuat dengan ❤️ oleh [jrevanaldi-ai](https://github.com/jrevanaldi-ai)**

⭐ **Jangan lupa beri bintang jika proyek ini membantu Anda!** ⭐

</div>
