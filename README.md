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
| 🗄️ | **SQLite Database** | Session disimpan lokal, aman dan persisten |
| 🎨 | **Formatted Output** | Pesan dengan format menarik dan mudah dibaca |
| ⚡ | **Multi-threading** | Handle multiple pesan secara concurrent |

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
# export GOWA_BOT_DB="/path/to/database.db"

# Log level (debug, info, warn, error)
# export GOWA_BOT_LOG_LEVEL="debug"
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

---

## 📜 Daftar Command

Bot ini menggunakan **prefix** `.` untuk command. Owner dapat menggunakan command tanpa prefix.

### Public Commands

| Command | Alias | Deskripsi | Contoh |
|---------|-------|-----------|--------|
| `.menu` | `.m`, `.h` | Tampilkan daftar command | `.menu` |
| `.ping` | `.p` | Cek latency bot | `.ping` |
| `.help` | `.info` | Lihat detail command | `.help ping` |

### Owner Only Commands

| Command | Alias | Deskripsi | Contoh |
|---------|-------|-----------|--------|
| `$` | `exec` | Eksekusi shell command | `$ls -la` |
| `.checkephemeral` | `.ce` | Cek status ephemeral group | `.checkephemeral` |

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
│ *MAIN:*
│   • help (info)
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
<summary><b>💻 Exec Command (Owner Only)</b></summary>

```
Kirim: $ls -la

Output:
✓ Output:
```
total 48
drwxr-xr-x  8 user user 4096 Mar 23 10:00 .
drwxr-xr-x 12 user user 4096 Mar 23 09:00 ..
-rw-r--r--  1 user user  234 Mar 23 10:00 .env
...
```
```

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
├── 📄 main.go              # Entry point aplikasi
├── 📦 go.mod               # Dependensi Go
├── 🔧 .env.example         # Template konfigurasi
├── 📖 README.md            # Dokumentasi (file ini)
│
├── 📂 client/
│   └── bot_client.go       # Wrapper WhatsApp client
│
├── 📂 commands/
│   ├── ping.go             # Command ping
│   ├── menu.go             # Command menu
│   ├── help.go             # Command help
│   ├── exec.go             # Command exec (owner)
│   ├── reply.go            # Helper reply message
│   └── checkephemeral.go   # Command debug ephemeral
│
├── 📂 helper/
│   ├── logger.go           # Logger dengan warna
│   ├── cache.go            # Sistem cache TTL
│   └── ephemeral.go        # Helper ephemeral message
│
└── 📂 lib/
    ├── types.go            # Tipe dasar & registry
    └── dispatcher.go       # Command dispatcher
```

---

## 🛠️ Development

### Menambah Command Baru

1. Buat file baru di folder `commands/`:

```go
// commands/halo.go
package commands

import "github.com/jrevanaldi-ai/gowa-bot/lib"

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
    _, err := ctx.SendMessage(createSimpleReply(message, ctx.MessageID, ctx.Sender.String()))
    return err
}
```

2. Daftarkan di `main.go`:

```go
func registerCommands(registry *lib.CommandRegistry) {
    // ... existing commands
    
    registry.Register(commands.HaloMetadata, commands.HaloHandler)
}
```

3. Build dan jalankan!

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
