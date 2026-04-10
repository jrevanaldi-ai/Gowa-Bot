# 🤖 Fitur Jadibot - Dokumentasi Lengkap

## 📖 Apa itu Jadibot?

**Jadibot** adalah fitur yang memungkinkan user membuat bot WhatsApp pribadi melalui bot induk. Setiap jadibot memiliki session terpisah dan memiliki semua fitur yang sama dengan bot induk.

## 🎯 Fitur Utama

| Fitur | Deskripsi |
|-------|-----------|
| **Pairing Code** | User mendapat pairing code untuk menghubungkan WhatsApp mereka |
| **Session Terpisah** | Setiap jadibot punya session sendiri dengan UUID unik |
| **Auto-Resume** | Jadibot otomatis resume saat bot induk restart |
| **Multi-User** | Banyak user bisa membuat jadibot secara bersamaan |
| **Full Features** | Jadibot memiliki semua command yang sama dengan bot induk |

## 📋 Command Jadibot

### 1. `.jadibot` - Buat Bot Baru

**Alias:** `.jb`, `.botbaru`

**Deskripsi:** Membuat bot WhatsApp baru dengan pairing code

**Usage:**
```
.jadibot <nomor_telepon>
```

**Contoh:**
```
.jadibot 6281234567890
```

**Flow:**
1. User kirim command dengan nomor telepon
2. Bot memvalidasi nomor
3. Bot membuat session baru dengan UUID unik
4. Bot generate pairing code
5. Bot kirim pairing code ke user
6. User pair di WhatsApp mereka
7. Jadibot aktif dan siap digunakan

**Output:**
```
✅ Jadibot Berhasil Dibuat!

┌─⦿ Info Jadibot
│ • ID: `abc123-def456-ghi789`
│ • Nomor: 6281234567890
│ • Pairing Code: `ABCD1234`
└──────────────

📝 Cara Pairing:
1. Buka WhatsApp di HP Anda
2. Menu → Perangkat Tertaut
3. Tautkan Perangkat
4. Masukkan pairing code di atas

⚠️ Penting:
• Pairing code kadaluarsa dalam 160 detik
• Gunakan jadibot dengan bijak
• Bot memiliki fitur yang sama dengan bot induk
```

---

### 2. `.listjadibot` - Lihat Daftar Jadibot

**Alias:** `.listjb`, `.jadibotlist`

**Deskripsi:** Melihat semua jadibot yang telah dibuat user

**Usage:**
```
.listjadibot
```

**Output:**
```
📋 Daftar Jadibot Anda

┌─⦿ Total: 2 bot

1. Jadibot ID: `abc123-def456-ghi789`
   • Nomor: 6281234567890
   • Status: 🟢 Aktif
   • Uptime: 02:15:30
   • Command:
     - `.stopjadibot abc123-def456-ghi789`
     - `.pausejadibot abc123-def456-ghi789`

2. Jadibot ID: `xyz789-uvw456-rst123`
   • Nomor: 6289876543210
   • Status: ⏸️ Paused
   • Command:
     - `.resumejadibot xyz789-uvw456-rst123`
     - `.stopjadibot xyz789-uvw456-rst123`

└──────────────

┌─⦿ Statistik
│ • 🟢 Aktif: 1
│ • ⏸️ Paused: 1
│ • ⏹️ Stopped: 0
└──────────────
```

---

### 3. `.stopjadibot` - Hentikan Jadibot

**Alias:** `.stopjb`, `.matibot`

**Deskripsi:** Menghentikan jadibot yang sedang berjalan

**Usage:**
```
.stopjadibot <id_jadibot>
```

**Contoh:**
```
.stopjadibot abc123-def456-ghi789
```

**Output:**
```
✅ Jadibot Berhasil Dihentikan!

┌─⦿ Info
│ • ID: `abc123-def456-ghi789`
│ • Nomor: 6281234567890
│ • Status: ⏹️ Stopped
└──────────────

📝 Command yang tersedia:
• `.resumejadibot abc123-def456-ghi789` - Resume jadibot
• `.listjadibot` - Lihat status jadibot

💡 Tips:
• Session jadibot tetap tersimpan
• Anda bisa resume kapan saja
• Jadibot tidak menggunakan resource saat stop
```

---

### 4. `.pausejadibot` - Pause Jadibot

**Alias:** `.pausejb`, `.jeda`

**Deskripsi:** Pause jadibot sementara (bisa diresume)

**Usage:**
```
.pausejadibot <id_jadibot>
```

**Contoh:**
```
.pausejadibot abc123-def456-ghi789
```

**Output:**
```
✅ Jadibot Berhasil Di-Pause!

┌─⦿ Info
│ • ID: `abc123-def456-ghi789`
│ • Nomor: 6281234567890
│ • Status: ⏸️ Paused
└──────────────

📝 Command yang tersedia:
• `.resumejadibot abc123-def456-ghi789` - Resume jadibot
• `.listjadibot` - Lihat status jadibot

💡 Tips:
• Jadibot pause tidak menggunakan resource
• Session tetap tersimpan
• Resume kapan saja saat dibutuhkan
```

---

### 5. `.resumejadibot` - Resume Jadibot

**Alias:** `.resumejb`, `.lanjut`

**Deskripsi:** Resume jadibot yang di-pause atau distop

**Usage:**
```
.resumejadibot <id_jadibot>
```

**Contoh:**
```
.resumejadibot abc123-def456-ghi789
```

**Output:**
```
✅ Jadibot Berhasil Di-Resume!

┌─⦿ Info
│ • ID: `abc123-def456-ghi789`
│ • Nomor: 6281234567890
│ • Status: 🟢 Aktif
└──────────────

Jadibot Anda sudah aktif kembali!

📝 Command yang tersedia:
• `.stopjadibot abc123-def456-ghi789` - Stop jadibot
• `.pausejadibot abc123-def456-ghi789` - Pause jadibot
• `.listjadibot` - Lihat status jadibot
```

---

## 🏗️ Arsitektur

```
Bot Induk (Main Bot)
    │
    ├── Database Manager (helper/database.go)
    │   └── SQLite: jadibots table
    │       • id (UUID)
    │       • owner_jid
    │       • phone_number
    │       • session_path
    │       • status (active/paused/stopped)
    │       • created_at, started_at, last_active_at
    │
    ├── Session Manager (helper/session_manager.go)
    │   ├── CreateJadibot()
    │   ├── StartJadibot() - Generate pairing code
    │   ├── StopJadibot()
    │   ├── PauseJadibot()
    │   ├── ResumeJadibot()
    │   ├── Auto-Reconnect
    │   └── Monitor Connection
    │
    ├── BotClient (client/bot_client.go)
    │   ├── JadibotSessionManager reference
    │   └── CommandContext dengan SessionManager
    │
    └── Commands (commands/jadibot/)
        ├── jadibot.go
        ├── listjadibot.go
        ├── stopjadibot.go
        └── pausejadibot.go
```

---

## 💾 Database Schema

```sql
CREATE TABLE jadibots (
    id TEXT PRIMARY KEY,              -- UUID unik
    owner_jid TEXT NOT NULL,          -- JID pemilik
    phone_number TEXT NOT NULL,       -- Nomor telepon
    session_path TEXT NOT NULL,       -- Path session folder
    status TEXT DEFAULT 'stopped',    -- active/paused/stopped
    created_at DATETIME,              -- Waktu dibuat
    started_at DATETIME,              -- Waktu mulai (aktif)
    last_active_at DATETIME           -- Waktu terakhir aktif
);

CREATE INDEX idx_jadibots_owner ON jadibots(owner_jid);
CREATE INDEX idx_jadibots_status ON jadibots(status);
```

---

## 🔄 Flow Lengkap

### 1. User Membuat Jadibot

```
User: .jadibot 6281234567890
  ↓
Bot: Validasi nomor telepon
  ↓
Bot: Cek apakah user sudah punya jadibot
  ↓
Bot: Generate UUID untuk session ID
  ↓
Bot: Buat session folder (sessions/jadibot_{uuid})
  ↓
Bot: Simpan ke database
  ↓
Bot: Create WhatsApp client
  ↓
Bot: Connect ke WhatsApp
  ↓
Bot: Generate pairing code
  ↓
User: Terima pairing code
  ↓
User: Pair di WhatsApp mereka
  ↓
Bot: Jadibot aktif ✓
```

### 2. Jadibot Berjalan

```
Jadibot Instance
    ↓
Event Handler (WhatsApp messages)
    ↓
Command Parser
    ↓
Command Registry
    ↓
Handler Execution
    ↓
Send Response
```

### 3. Auto-Resume saat Bot Induk Restart

```
Bot Induk Start
    ↓
Load Database
    ↓
Get Active Jadibots
    ↓
For each active jadibot:
    - Load session
    - Connect WhatsApp
    - Resume monitoring
    ↓
All Jadibots Active ✓
```

---

## 📂 Struktur File

```
gowa-bot/
├── helper/
│   ├── database.go           # Database manager
│   └── session_manager.go    # Session manager
│
├── commands/
│   └── jadibot/
│       ├── jadibot.go        # Command .jadibot
│       ├── listjadibot.go    # Command .listjadibot
│       ├── stopjadibot.go    # Command .stopjadibot
│       └── pausejadibot.go   # Command .pausejadibot & .resumejadibot
│
├── client/
│   └── bot_client.go         # Updated dengan SessionManager
│
├── lib/
│   └── types.go              # Updated dengan JadibotSessionManagerInterface
│
├── sessions/                 # Folder untuk session jadibot
│   └── jadibot_{uuid}/
│       └── jadibot.db
│
└── main.go                   # Updated dengan init SessionManager
```

---

## 🔐 Keamanan

### 1. Ownership Verification
- Setiap jadibot hanya bisa dikontrol oleh pembuatnya
- Verifikasi berdasarkan JID WhatsApp

### 2. Session Isolation
- Setiap jadibot punya session folder terpisah
- Session tidak saling tumpang tindih

### 3. Resource Management
- User hanya bisa membuat 1 jadibot
- Pause/Stop untuk hemat resource
- Auto-reconnect dengan limit

---

## ⚙️ Konfigurasi

Tidak ada konfigurasi tambahan untuk fitur jadibot. Semua settings sudah otomatis saat bot induk dijalankan.

**Yang perlu diperhatikan:**
- Pastikan ada storage cukup untuk session jadibot
- Monitor jumlah jadibot aktif (resource usage)
- Backup folder `sessions/` secara berkala

---

## 🚀 Cara Penggunaan

### 1. Jalankan Bot Induk

```bash
go run . -phone 6281234567890
```

### 2. User Membuat Jadibot

User kirim command ke bot induk:
```
.jadibot 6289876543210
```

### 3. User Terima Pairing Code

Bot akan mengirim pairing code:
```
✅ Jadibot Berhasil Dibuat!

┌─⦿ Info Jadibot
│ • ID: `abc123-def456-ghi789`
│ • Pairing Code: `ABCD1234`
└──────────────
```

### 4. User Pair di WhatsApp

User buka WhatsApp → Perangkat Tertaut → Tautkan → Masukkan pairing code

### 5. Jadibot Aktif

Setelah pairing berhasil, jadibot otomatis aktif dan siap digunakan!

---

## 💡 Tips & Trik

### Untuk User:

1. **Pause jika tidak digunakan** - Hemat resource server
2. **Cek status berkala** - Gunakan `.listjadibot`
3. **Jangan spam pairing** - Pairing code kadaluarsa dalam 160 detik
4. **Backup ID jadibot** - Untuk kontrol nanti

### Untuk Admin:

1. **Monitor active bots** - Cek resource usage
2. **Backup sessions** - Folder `sessions/` penting
3. **Set limit** - 1 user = 1 jadibot (sudah implemented)
4. **Monitor logs** - Cek log untuk troubleshooting

---

## 🐛 Troubleshooting

### 1. Jadibot tidak bisa resume

**Solusi:**
- Cek apakah session folder masih ada
- Cek log untuk error detail
- Coba buat jadibot baru jika session corrupt

### 2. Pairing code tidak bekerja

**Solusi:**
- Pastikan nomor telepon valid (format internasional)
- Tunggu beberapa detik dan coba lagi
- Pairing code kadaluarsa dalam 160 detik

### 3. Jadibot tidak merespon

**Solusi:**
- Cek status dengan `.listjadibot`
- Jika status "active" tapi tidak merespon, coba pause lalu resume
- Cek koneksi WhatsApp di log

---

## 📊 Limitasi

| Limitasi | Value |
|----------|-------|
| Max jadibot per user | 1 bot |
| Pairing code expiry | 160 detik |
| Auto-reconnect attempts | 5 kali |
| Reconnect delay | 5-25 detik (exponential) |

---

## 🎯 Future Improvements

Berikut beberapa ide improvement untuk masa depan:

- [ ] Multi jadibot per user (dengan limit)
- [ ] Dashboard web untuk manage jadibot
- [ ] Custom command per jadibot
- [ ] Rate limiting per jadibot
- [ ] Statistics & analytics
- [ ] Export/import session
- [ ] Auto-delete inactive jadibot
- [ ] Premium features untuk jadibot

---

## 📝 Catatan Penting

1. **Setiap jadibot adalah WhatsApp client terpisah** - Memerlukan resource (RAM, CPU)
2. **Session disimpan lokal** - Backup folder `sessions/` secara berkala
3. **Pairing code sekali pakai** - Jika gagal, harus buat jadibot baru
4. **Jadibot = Bot Induk** - Fitur yang sama, tidak ada perbedaan

---

## ✨ Kesimpulan

Fitur **Jadibot** memungkinkan user membuat bot WhatsApp pribadi melalui bot induk dengan mudah. Setiap jadibot memiliki:

- ✅ Session terpisah dengan UUID unik
- ✅ Pairing code untuk koneksi
- ✅ Semua fitur yang sama dengan bot induk
- ✅ Kontrol penuh (stop, pause, resume)
- ✅ Auto-resume saat bot induk restart
- ✅ Security dengan ownership verification

**Selamat menggunakan fitur Jadibot!** 🎉
