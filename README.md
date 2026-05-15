<div align="center">

# 🤖 Gowa-Bot

**WhatsApp Bot sederhana dan powerful yang dibangun dengan Go**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
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
| ⬇️ | **Downloader** | Download dari YouTube, Spotify, Instagram, TikTok, GitHub |
| 🚫 | **Ban System** | Ban user atau group dari pemakaian bot |
| 🛡️ | **Eval Sandbox** | Eksekusi kode Go runtime via yaegi (owner only) |

---

## 📋 Prasyarat

Sebelum memulai, pastikan Anda telah menginstal:

- **[Go](https://go.dev/dl/)** versi 1.21 atau lebih tinggi
- **[Git](https://git-scm.com/downloads)** untuk clone repository
- **WhatsApp** aktif untuk pairing bot

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

> 💡 Semua flag punya fallback ke environment variable dengan prefix `GOWA_BOT_`.

---

## 📜 Daftar Command

Bot menggunakan **prefix** `.` untuk command (bisa diganti dengan `.setprefix`). Owner dapat menggunakan command **tanpa prefix**.

### 🛠️ Utility

| Command | Alias | Deskripsi | Contoh |
|---------|-------|-----------|--------|
| `.ping` | `.p` | Cek latency bot | `.ping` |

### 📋 General

| Command | Alias | Deskripsi | Contoh |
|---------|-------|-----------|--------|
| `.menu` | `.m`, `.h` | Tampilkan daftar command | `.menu` |
| `.help` | `.info` | Lihat detail command | `.help ping` |
| `.getpp` | `.pp` | Ambil foto profil user | `.getpp @user` |
| `.donasi` | `.donate` | Buat QRIS donasi (butuh MustikaPay) | `.donasi 10000` |
| `.cekdonasi` | `.checkdonate` | Cek status donasi | `.cekdonasi <refno>` |
| `.lune` | - | Tanya AI Claude | `.lune apa itu Go?` |

> 💬 **Keyword `lune`:** Pesan apa pun yang mengandung kata "lune" (tanpa prefix sekalipun) akan otomatis dijawab oleh AI. Contoh: `"hai lune apa kabar?"`

### ⬇️ Download

| Command | Alias | Deskripsi | Contoh |
|---------|-------|-----------|--------|
| `.play` | - | Download audio dari YouTube | `.play lagu galau` |
| `.spotify` | - | Download dari Spotify | `.spotify <url>` |
| `.instagram` | - | Download Instagram post/reel | `.instagram <url>` |
| `.tiktok` | - | Download video TikTok | `.tiktok <url>` |
| `.ttsearch` | - | Search video TikTok | `.ttsearch keyword` |
| `.github` | - | Info / download repo GitHub | `.github user/repo` |

### 🔐 Owner Only

| Command | Alias | Deskripsi | Contoh |
|---------|-------|-----------|--------|
| `$` | `exec` | Eksekusi shell command | `$ls -la` |
| `>>` | `ev`, `eval` | Eksekusi kode Go runtime (yaegi) | `>> return ctx.Chat.String()` |
| `.setmode` | `.mode` | Ganti mode bot (self/public) | `.setmode self` |
| `.setprefix` | - | Ganti prefix command | `.setprefix !` |
| `.infoserver` | - | Info server (CPU/RAM/uptime) | `.infoserver` |
| `.react` | - | Reaksi emoji ke pesan | `.react ❤️` |
| `.bangroup` | - | Ban group dari pemakaian bot | `.bangroup` |
| `.unbangroup` | - | Unban group | `.unbangroup` |
| `.banuser` | - | Ban user | `.banuser @user` |
| `.unbanuser` | - | Unban user | `.unbanuser @user` |

### 🤝 Jadibot (Multi-Bot)

Memungkinkan user lain pairing nomor mereka sebagai sub-bot di bawah Gowa-Bot utama.

| Command | Alias | Deskripsi | Contoh |
|---------|-------|-----------|--------|
| `.jadibot` | - | Daftar nomor jadi sub-bot | `.jadibot 6281234567890` |
| `.listjadibot` | - | List semua jadibot aktif | `.listjadibot` |
| `.stopjadibot` | - | Hentikan jadibot | `.stopjadibot <id>` |
| `.pausejadibot` | - | Pause jadibot (bisa di-resume) | `.pausejadibot <id>` |
| `.resumejadibot` | - | Resume jadibot yang di-pause | `.resumejadibot <id>` |
| `.removejadibot` | - | Hapus jadibot (permanen) | `.removejadibot <id>` |

### 🐞 Debug

| Command | Alias | Deskripsi | Contoh |
|---------|-------|-----------|--------|
| `.checkephemeral` | `.ce` | Cek status ephemeral group | `.checkephemeral` |

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
╭───⦿ GOWA-BOT ⦿───
│
│ *UTILITY:*
│   • ping (p)
│
│ *GENERAL:*
│   • help (info)
│   • getpp (pp)
│   • donasi (donate)
│   • lune
│
│ *DOWNLOAD:*
│   • play
│   • spotify
│   • instagram
│   • tiktok
│
╰────────────────
```

</details>

<details>
<summary><b>🏓 Ping Command</b></summary>

```
Kirim: .ping

Output:
🏓 Pong!

┌─⦿ Info Bot
│ • Latency: 45 ms
│ • Status: Online ✓
│ • Uptime: 00:15:32
└──────────────
```

</details>

<details>
<summary><b>🧠 AI Lune</b></summary>

```
Kirim: .lune jelaskan apa itu goroutine

atau cukup:
Kirim: hey lune, apa kabar?

Output:
[Respon AI Claude]
```

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
╭──⦿ HELP: PING ⦿
│
│  Category: Utility
│  Description: Cek respon bot dan latency
│  Command: .ping
│  Aliases: .p
│  Example: .ping
│  Access: Public
│
╰──────────────────────
```

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
├── 📄 main.go              # Entry point — daftarkan command di sini
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
│   ├── 📂 owner/             # exec, eval, setmode, setprefix, ban, dll
│   ├── 📂 jadibot/           # multi-bot management
│   ├── 📂 download/          # play, spotify, instagram, tiktok, github
│   └── 📂 debug/             # checkephemeral
│
├── 📂 helper/
│   ├── logger.go             # Logger berwarna
│   ├── cache.go              # Cache TTL in-memory
│   ├── ephemeral.go          # Helper pesan ephemeral
│   ├── message.go            # Builder reply message
│   ├── database.go           # SQLite manager (jadibot/banned/donasi)
│   ├── session_manager.go    # Jadibot session manager
│   ├── ai.go                 # Claude AI service
│   └── mustikapay.go         # MustikaPay QRIS integration
│
├── 📂 lib/
│   ├── types.go              # CommandRegistry, CommandContext, interfaces
│   └── dispatcher.go         # (legacy, tidak dipakai di flow saat ini)
│
├── 📂 gowa-lib/              # Fork lokal library Gowa
└── 📂 sessions/              # Storage untuk session jadibot (auto-generated)
```

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

2. **Wajib** daftarkan di `main.go::registerCommands()`:

```go
func registerCommands(registry *lib.CommandRegistry) {
    // ... existing commands
    registry.Register(utility.HaloMetadata, utility.HaloHandler)
}
```

> ⚠️ **Tidak ada autoload.** Kalau lupa daftar di sini, command tidak akan jalan.

3. Build dan jalankan!

### Arsitektur Singkat

Alur pesan masuk:

```
WhatsApp Event
    ↓
BotClient.EventHandler (client/bot_client.go)
    ↓
processMessage()
    ├─ Cek self-mode & IsFromMe
    ├─ Ekstrak teks dari semua tipe pesan
    ├─ Unwrap edited/ephemeral/view-once
    ├─ Cek ban (skip kalau owner)
    ├─ Parse command (prefix / $ / owner-tanpa-prefix / keyword "lune")
    ├─ Validasi OwnerOnly
    └─ Jalankan handler dengan CommandContext
```

Untuk detail teknis lebih lanjut, lihat **[CLAUDE.md](CLAUDE.md)**.

### Tips Development

- **Toolchain Go:** `go.mod` declare `go 1.26` tapi project tested di 1.21+. Sesuaikan dengan environment Anda.
- **Fork Gowa:** Direktori `gowa-lib/` adalah fork lokal — edit di sana langsung mempengaruhi bot.
- **Session DB:** Jangan hapus `gowa-bot.db` saat bot jalan, session WhatsApp akan hilang.
- **Cyclic import:** Kalau bikin command yang butuh akses ke `BotClient`, pakai `lib.BotClientInterface`, jangan import `client/` langsung.

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
