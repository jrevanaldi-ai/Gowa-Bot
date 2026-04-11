# Fitur Donasi - MustikaPay Integration

## 📋 Ringkasan

Fitur donasi memungkinkan pengguna untuk melakukan donasi ke bot menggunakan QRIS (Quick Response Code Indonesian Standard) melalui payment gateway MustikaPay.

## 🎯 Command Metadata

### Command `.donasi`
- **Command**: `.donasi`
- **Tag**: `main`
- **Aliases**: `.donate`, `.qr`
- **Deskripsi**: Buat QRIS payment untuk donasi
- **Contoh**: `.donasi 10000`
- **Akses**: Public (semua user bisa menggunakan)

### Command `.cekdonasi`
- **Command**: `.cekdonasi`
- **Tag**: `main`
- **Aliases**: `.cekdon`, `.statusdonasi`
- **Deskripsi**: Cek status donasi via Ref No atau lihat riwayat
- **Contoh**: `.cekdonasi QR12345` atau `.cekdonasi` (tanpa argumen)
- **Akses**: Public (semua user bisa menggunakan)

## 📖 Cara Penggunaan

### 1. **Buat Donasi Baru**
Kirim command dengan nominal yang diinginkan:
```
.donasi 10000
```

Bot akan merespon dengan:
- QR Code image untuk discan
- Ref No untuk tracking
- Detail pembayaran
- Instruksi pembayaran

### 2. **Cek Status Donasi**
Ada 2 cara:

**Via Ref No:**
```
.cekdonasi QR12345
```

**Lihat Riwayat Donasi:**
```
.cekdonasi
```
(tanpa argumen akan menampilkan riwayat donasi user)

## 🔧 Implementasi Teknis

### File yang Dibuat/Dimodifikasi

#### 1. **File Baru: `/helper/mustikapay.go`**
API client untuk MustikaPay payment gateway.

**Functions:**
- `NewMustikaPayClient(apiKey, logger)`: Create new API client
- `CreateQRIS(req)`: Create QRIS payment
- `CheckPaymentStatus(refNo)`: Check payment status
- `WaitForPayment(refNo, interval, maxAttempts)`: Polling payment status

**Request/Response Format:**

```go
// Create Payment Request
type CreatePaymentRequest struct {
    Amount       int    `form:"amount"`        // Required
    ProductName  string `form:"product_name"`  // Optional
    CustomerName string `form:"customer_name"` // Optional
    RedirectURL  string `form:"redirect_url"`  // Optional
}

// Create Payment Response
type CreatePaymentResponse struct {
    Status      json.RawMessage `json:"status"`  // Bisa string atau bool
    Message     string          `json:"message"`
    RefNo       string          `json:"ref_no"`
    QRString    string          `json:"qr_string"`
    QRImageURL  string          `json:"qr_image"`
    Amount      string          `json:"amount"`
    ProductName string          `json:"product_name"`
    ExpiresAt   string          `json:"expires_at"`
    Type        string          `json:"type"`
}

// Check Payment Response
type CheckPaymentResponse struct {
    RefNo  string `json:"ref_no"`
    Status string `json:"status"`    // success/pending/expired/failed
    Type   string `json:"type"`      // QRIS/VA/Retail
    Amount int    `json:"amount"`
    Issuer string `json:"issuer"`    // Gopay/OVO/Dana/etc
    Payor  string `json:"payor"`
}
```

**API Integration:**
- **Base URL**: `https://mustikapayment.com`
- **Create Endpoint**: `POST /api/createpay`
- **Check Endpoint**: `GET /api/cekpay?ref_no={ref_no}`
- **Auth Header**: `X-Api-Key: MP-XXXXXX`
- **Content-Type**: `application/x-www-form-urlencoded`
- **MDR Fee**: 0.7%

#### 2. **File Baru: `/commands/general/donasi.go`**
Handler untuk command donasi dan cek status.

**Functions:**
- `DonasiHandler(ctx)`: Main handler untuk create QRIS payment
- `CekDonasiHandler(ctx)`: Handler untuk cek status donasi
- `monitorDonation()`: Background polling untuk detect payment
- `sendSuccessNotification()`: Kirim notifikasi pembayaran sukses
- `sendExpiredNotification()`: Kirim notifikasi QRIS expired
- `showDonationFromDatabase()`: Tampilkan detail donasi dari DB
- `showUserDonationHistory()`: Tampilkan riwayat donasi user

**Payment Flow:**
1. User kirim `.donasi 10000`
2. Bot validate amount (min 1000, max 10000000)
3. Bot call MustikaPay API create QRIS
4. Bot save donation info to database (status: pending)
5. Bot send QR image + caption ke user
6. Bot start background polling (max 90 attempts, interval 10s = 15 menit)
7. Saat payment detected (status: success), bot send notification
8. Jika expired/failed, bot send notification juga

#### 3. **File Modifikasi: `/helper/database.go`**

**Penambahan Tabel `donations`:**
```sql
CREATE TABLE IF NOT EXISTS donations (
    id TEXT PRIMARY KEY,
    ref_no TEXT NOT NULL UNIQUE,
    user_jid TEXT NOT NULL,
    user_name TEXT,
    amount INTEGER NOT NULL,
    qr_string TEXT,
    qr_image_url TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    payment_type TEXT,
    issuer TEXT,
    payor TEXT,
    product_name TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_donations_user ON donations(user_jid);
CREATE INDEX IF NOT EXISTS idx_donations_ref ON donations(ref_no);
CREATE INDEX IF NOT EXISTS idx_donations_status ON donations(status);
```

**Penambahan Functions:**
```go
type DonationInfo struct {
    ID            string
    RefNo         string
    UserJID       string
    UserName      string
    Amount        int
    QRString      string
    QRImageURL    string
    Status        string       // pending/success/expired/failed
    PaymentType   string       // QRIS/VA/Retail
    Issuer        string       // Gopay/OVO/Dana
    Payor         string
    ProductName   string
    CreatedAt     interface{}
    UpdatedAt     interface{}
}

// Database functions
- CreateDonation(info)
- GetDonation(refNo)
- UpdateDonationStatus(refNo, status, paymentType, issuer, payor)
- GetUserDonations(userJID)
- GetPendingDonations()
```

#### 4. **File Modifikasi: `/main.go`**

**Penambahan Flag:**
```go
mustikaPayAPIKey = flag.String("mustika-api-key", "", "MustikaPay API Key")
```

**Environment Variable:**
```go
// Read from env if not provided via flag
if *mustikaPayAPIKey == "" {
    *mustikaPayAPIKey = os.Getenv("GOWA_BOT_MUSTIKA_API_KEY")
}

// Initialize if configured
if *mustikaPayAPIKey != "" {
    general.SetMustikaPayAPIKey(*mustikaPayAPIKey)
    logger.Info("MustikaPay payment integration enabled")
}
```

**Command Registration:**
```go
registry.Register(general.DonasiMetadata, general.DonasiHandler)
registry.Register(general.CekDonasiMetadata, general.CekDonasiHandler)
```

## 💰 Amount Validation

| Rule | Value |
|------|-------|
| Minimal | Rp 1.000 |
| Maksimal | Rp 10.000.000 |
| MDR Fee | 0.7% (ditanggung bot owner) |

## ⏱️ Payment Timeout & Polling

| Parameter | Value |
|-----------|-------|
| Polling Interval | 10 detik |
| Max Attempts | 90 kali |
| Total Timeout | 15 menit (900 detik) |
| Auto-update DB | ✅ Ya, saat status berubah |

## 🔔 Notifikasi Otomatis

Bot akan mengirim notifikasi otomatis saat:

1. **Payment Success** ✅
   - Ref No
   - Nominal
   - Metode pembayaran
   - Nama pembayar
   - Thank you message

2. **Payment Expired** ⏰
   - Ref No
   - Nominal
   - Instructions untuk buat baru

3. **Payment Timeout** ⏰
   - Ref No
   - Nominal
   - Instructions untuk buat baru

## 📊 Contoh Output

### Command `.donasi 10000`

**Loading Message:**
```
⏳ *Membuat QRIS payment...*

┌─⦿ *Info*
│ • Nominal: Rp 10.000
│ • Status: Memproses...
└──────────────

_Mohon tunggu..._
```

**Success Response (QR Image + Caption):**
```
✅ *QRIS Payment Berhasil Dibuat!*

┌─⦿ *Detail Pembayaran*
│ • *Ref No:* `QR12345`
│ • *Nominal:* Rp 10.000
│ • *Produk:* Donasi Gowa-Bot
└──────────────

📱 *Cara Pembayaran:*
1. 📸 Scan QR Code di atas
2. 💰 Masukkan nominal: Rp 10.000
3. ✅ Konfirmasi pembayaran

⏰ *Batas Waktu:* 15 menit
🔍 *Cek Status:* `.cekdonasi QR12345`

_Bot akan otomatis mendeteksi pembayaran Anda!_
```

**Payment Success Notification:**
```
🎉 *Pembayaran Berhasil!*

┌─⦿ *Detail*
│ • *Ref No:* `QR12345`
│ • *Nominal:* Rp 10.000
│ • *Metode:* QRIS
│ • *Pembayar:* John Doe
│ • *Status:* ✅ SUCCESS
└──────────────

_Terima kasih atas donasi Anda! 🙏_
_Dukungan Anda sangat berarti untuk pengembangan bot ini._
```

### Command `.cekdonasi QR12345`

**Pending Status:**
```
⏳ *Status Donasi*

┌─⦿ *Detail*
│ • *Ref No:* `QR12345`
│ • *Nominal:* Rp 10.000
│ • *Produk:* Donasi Gowa-Bot
│ • *Status:* ⏳ PENDING
└──────────────

💡 *Tips:*
• Bot akan otomatis mendeteksi pembayaran
• QRIS berlaku selama 15 menit
• Cek kembali dalam beberapa saat
```

**Success Status:**
```
✅ *Status Donasi*

┌─⦿ *Detail*
│ • *Ref No:* `QR12345`
│ • *Nominal:* Rp 10.000
│ • *Produk:* Donasi Gowa-Bot
│ • *Status:* ✅ SUCCESS
│ • *Metode:* QRIS
│ • *Issuer:* GOPAY
│ • *Pembayar:* John Doe
└──────────────
```

### Command `.cekdonasi` (Riwayat)

```
📋 *Riwayat Donasi Anda*

┌─⦿ *Total: 5 donasi*

1. ✅ Rp 100.000 - QR12345
2. ✅ Rp 50.000 - QR12346
3. ⏳ Rp 10.000 - QR12347
4. ❌ Rp 5.000 - QR12348
5. ✅ Rp 25.000 - QR12349
└──────────────

┌─⦿ *Statistik*
│ • ✅ Sukses: 3
│ • ⏳ Pending: 1
│ • 💰 Total Donasi: Rp 175.000
└──────────────

📝 Command:
• `.cekdonasi <ref_no>` - Cek detail donasi
• `.donasi <nominal>` - Buat donasi baru
```

## 🔐 Security Considerations

1. **API Key Protection:**
   - API key disimpan di memory
   - Tidak di-hardcode di source code
   - Bisa via flag atau environment variable

2. **User Validation:**
   - User hanya bisa lihat donasi milik sendiri
   - Access control di `CekDonasiHandler`

3. **Amount Validation:**
   - Minimal dan maksimal amount
   - Prevent integer overflow

4. **Database Persistence:**
   - Semua transaksi disimpan ke database
   - Bisa di-audit dan di-trace

## 🚀 Potential Improvements

1. **Webhook Integration:**
   - Terima webhook dari MustikaPay (real-time notification)
   - Kurangi reliance pada polling

2. **Payment Methods:**
   - Support Virtual Account
   - Support Retail (Alfamart/Indomaret)
   - Auto-detect dari response API

3. **Donation Leaderboard:**
   - Top donatur bulanan
   - Reward system untuk donatur tetap

4. **Auto-retry Failed Payment:**
   - Auto-create QRIS baru jika expired
   - Notify user dengan quick command

5. **Donation Goals:**
   - Target donasi bulanan
   - Progress bar visualization

6. **Receipt Generation:**
   - Generate receipt image
   - Shareable ke social media

## 🐛 Troubleshooting

### API Key Not Configured
**Error:** "Fitur donasi belum dikonfigurasi!"

**Solution:**
```bash
# Via flag
./gowa-bot -mustika-api-key "MP-XXXXXX"

# Via environment
export GOWA_BOT_MUSTIKA_API_KEY="MP-XXXXXX"
./gowa-bot
```

### Payment Not Detected
**Issue:** Bot tidak detect pembayaran yang sudah sukses

**Possible Causes:**
1. API key invalid
2. Network issue ke MustikaPay
3. Polling timeout (15 menit)

**Solution:**
- Cek log bot untuk error messages
- Manual cek status via `.cekdonasi <ref_no>`
- Pastikan pembayaran dilakukan dalam 15 menit

### Database Error
**Error:** "Failed to save donation to database"

**Solution:**
- Cek permission write ke folder database
- Pastikan database file tidak corrupt
- Restart bot untuk recreate connection

## ✅ Status

**STATUS: IMPLEMENTED** ✅

Fitur donasi sudah sepenuhnya diimplementasikan dan siap digunakan setelah:
1. Setup MustikaPay API key
2. Build bot dengan `go build`
3. Run bot dengan flag atau environment variable

## 📝 Setup Checklist

- [ ] Daftar akun di https://mustikapayment.com
- [ ] Get API Key dari dashboard
- [ ] Set environment variable `GOWA_BOT_MUSTIKA_API_KEY`
- [ ] Build bot: `go build -o gowa-bot`
- [ ] Run bot: `./gowa-bot -phone 628xxx`
- [ ] Test command: `.donasi 10000`
- [ ] Verify QR image muncul
- [ ] Test payment (small amount)
- [ ] Verify auto-detection works
- [ ] Test `.cekdonasi` command
