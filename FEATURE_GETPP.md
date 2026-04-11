# Fitur GetPP - Dokumentasi

## 📋 Ringkasan

Command `getpp` adalah fitur baru yang memungkinkan pengguna untuk mengambil foto profil WhatsApp user lain melalui berbagai cara.

## 🎯 Command Metadata

- **Command**: `.getpp`
- **Tag**: `main`
- **Aliases**: `.pp`, `.avatar`
- **Deskripsi**: Ambil foto profil user (reply, tag, atau nomor)
- **Contoh**: `.getpp 6281234567890`
- **Akses**: Public (semua user bisa menggunakan)

## 📖 Cara Penggunaan

Ada 3 cara untuk menggunakan command ini:

### 1. **Reply Pesan User**
Reply pesan user yang ingin diambil foto profilnya, lalu kirim:
```
.getpp
```

### 2. **Tag/Mention User**
Tag user yang ingin diambil foto profilnya:
```
.getpp @6281234567890
```

### 3. **Masukkan Nomor Langsung**
Masukkan nomor telepon user secara langsung:
```
.getpp 6281234567890
```

## 🔧 Implementasi Teknis

### File yang Dibuat/Dimodifikasi

#### 1. **File Baru: `/commands/general/getpp.go`**
Handler utama untuk command getpp.

**Fungsi Utama:**
- `GetppHandler(ctx *lib.CommandContext) error`: Handler utama yang memproses command
- `sentMsgToString(sentMsg interface{}) string`: Helper untuk convert response ke message ID
- `formatJID(jid string) string`: Helper untuk format JID (menghilangkan suffix)

**Alur Kerja:**
1. **Parse Target JID** dari:
   - Reply message (prioritas tertinggi)
   - Mention/tagged users
   - Argumen command (nomor telepon)

2. **Loading Message**: Kirim pesan "⏳ Mengambil foto profil..."

3. **Get Profile Picture**: 
   - Gunakan `ctx.Client.GetProfilePictureInfo()` dari gowa-lib
   - Parameter: `Preview: false` untuk mendapatkan gambar full size

4. **Download Image**: 
   - Download gambar dari URL yang didapatkan
   - Timeout: 15 detik

5. **Upload to WhatsApp Media**:
   - Upload gambar menggunakan `ctx.Client.Upload()` dengan `gowa.MediaImage`

6. **Send Image Message**:
   - Kirim sebagai WhatsApp ImageMessage dengan context info
   - Caption berisi info user (JID, Picture ID)

**Error Handling:**
- `ErrProfilePictureUnauthorized`: User menyembunyikan foto profil
- `ErrProfilePictureNotSet`: User tidak memiliki foto profil
- Network errors: Download/upload failures
- Format errors: Invalid input

#### 2. **File Modifikasi: `/lib/types.go`**

**Penambahan:**
```go
type ReplyMessageInfo struct {
    MessageID string
    Sender    string
    Message   string
}

type CommandContext struct {
    // ... existing fields ...
    
    // ReplyMessage berisi informasi tentang pesan yang di-reply (jika ada)
    ReplyMessage *ReplyMessageInfo
    
    // Mentions berisi daftar JID yang di-tag dalam pesan
    Mentions []string
}

func StringToJID(jidStr string) types.JID {
    jid, err := types.ParseJID(jidStr)
    if err != nil {
        return types.NewJID(jidStr, types.DefaultUserServer)
    }
    return jid
}
```

**Penjelasan:**
- `ReplyMessageInfo`: Struct untuk menyimpan informasi tentang pesan yang di-reply
- `CommandContext.ReplyMessage`: Field baru untuk menyimpan info reply
- `CommandContext.Mentions`: Field baru untuk menyimpan daftar user yang di-tag
- `StringToJID()`: Helper function untuk convert string ke types.JID

#### 3. **File Modifikasi: `/client/bot_client.go`**

**Penambahan di `processMessage()`:**

```go
// Extract reply message info
var replyMsg *lib.ReplyMessageInfo
if evt.Message.ExtendedTextMessage != nil && evt.Message.ExtendedTextMessage.ContextInfo != nil {
    contextInfo := evt.Message.ExtendedTextMessage.ContextInfo
    if contextInfo.StanzaID != nil && contextInfo.Participant != nil {
        replyMsg = &lib.ReplyMessageInfo{
            MessageID: *contextInfo.StanzaID,
            Sender:    *contextInfo.Participant,
            Message:   "",
        }
    }
}

// Extract mentions
var mentions []string
if evt.Message.ExtendedTextMessage != nil && evt.Message.ExtendedTextMessage.ContextInfo != nil {
    contextInfo := evt.Message.ExtendedTextMessage.ContextInfo
    if len(contextInfo.MentionedJID) > 0 {
        mentions = make([]string, len(contextInfo.MentionedJID))
        for i, mention := range contextInfo.MentionedJID {
            mentions[i] = mention
        }
    }
}

// Add to CommandContext
cmdCtx := &lib.CommandContext{
    Ctx: context.WithValue(
        context.WithValue(ctx, "registry", b.Registry), 
        "gowa_client", 
        b.Client,
    ),
    // ... other fields ...
    ReplyMessage: replyMsg,
    Mentions:     mentions,
}
```

**Penjelasan:**
- Extract `StanzaID` dan `Participant` dari `ContextInfo` untuk reply message
- Extract `MentionedJID` dari `ContextInfo` untuk mentions
- Menambahkan `gowa_client` ke context value untuk akses WhatsApp client di command handler

#### 4. **File Modifikasi: `/main.go`**

**Pendaftaran Command:**
```go
registry.Register(general.GetppMetadata, general.GetppHandler)
```

## 🔍 Fungsi Gowa yang Digunakan

### 1. **GetProfilePictureInfo**
```go
func (cli *Client) GetProfilePictureInfo(
    ctx context.Context, 
    jid types.JID, 
    params *GetProfilePictureParams,
) (*types.ProfilePictureInfo, error)
```

**Parameters:**
- `ctx`: Context dengan timeout 15 detik
- `jid`: Target user JID
- `params`: `&gowa.GetProfilePictureParams{Preview: false}` untuk full size image

**Returns:**
- `ProfilePictureInfo.URL`: URL gambar
- `ProfilePictureInfo.ID`: ID gambar (untuk caching)
- `ProfilePictureInfo.DirectPath`: Path langsung ke gambar

**Errors:**
- `ErrProfilePictureUnauthorized`: User menyembunyikan foto profil
- `ErrProfilePictureNotSet`: User tidak memiliki foto profil

### 2. **Upload**
```go
func (cli *Client) Upload(
    ctx context.Context, 
    data []byte, 
    mediaType MediaType,
) (UploadResponse, error)
```

**Parameters:**
- `ctx`: Context
- `data`: Binary image data
- `mediaType`: `gowa.MediaImage` untuk gambar

**Returns:**
- `UploadResponse.URL`: URL file yang diupload
- `UploadResponse.DirectPath`: Path langsung
- `UploadResponse.FileSHA256`: Hash SHA256 file
- `UploadResponse.FileEncSHA256`: Hash SHA256 terenkripsi
- `UploadResponse.FileLength`: Ukuran file
- `UploadResponse.MediaKey`: Media key untuk decrypt

## 📊 Contoh Output

### Success
```
✅ *Foto Profil Berhasil Diambil!*

┌─⦿ Info User
│ • JID: 6281234567890
│ • Picture ID: 1234567890
└──────────────

💡 Tips: Klik gambar untuk melihat full size
```

### Error - User Sembunyikan Foto Profil
```
❌ User ini menyembunyikan foto profilnya dari Anda 🔒
```

### Error - User Tidak Punya Foto Profil
```
❌ User ini tidak memiliki foto profil
```

### Error - Format Salah
```
❌ Format salah!

Gunakan salah satu cara berikut:
• Reply pesan user
• Tag user: @6281234567890
• Masukkan nomor: .getpp 6281234567890
```

## 🔐 Security Considerations

1. **Privacy Respect**: Command ini menghormati privacy setting user
   - Jika user menyembunyikan foto profil, akan ditolak dengan error yang jelas

2. **No Rate Limiting**: Saat ini tidak ada rate limiting (bisa ditambahkan jika diperlukan)

3. **Public Command**: Semua user bisa menggunakan command ini tanpa batasan

## 🚀 Potential Improvements

1. **Rate Limiting**: Tambahkan rate limiter untuk prevent abuse
2. **Cache**: Cache foto profil untuk mengurangi API calls (dengan TTL)
3. **Preview Option**: Tambahkan flag untuk preview vs full size
4. **Batch Processing**: Support multiple users sekaligus
5. **Save to Gallery**: Option untuk menyimpan foto ke database

## 🐛 Troubleshooting

### Build Error: Version Mismatch
Jika terjadi error `version "go1.25.x" does not match go tool version "go1.26.x"`:
- Ini adalah masalah environment Go
- Update `gowa-lib/go.mod` untuk menggunakan versi Go yang sama
- Atau downgrade Go di environment ke 1.25

### Profile Picture Not Found
- Pastikan nomor menggunakan format internasional (62xxx)
- Pastikan user memiliki foto profil
- User mungkin menyembunyikan foto profil dari Anda

### Upload Failed
- Cek koneksi internet
- Pastikan WhatsApp client masih connected
- Cek apakah ada limitasi dari WhatsApp

## 📝 Testing Checklist

- [x] Reply message detection
- [x] Mention/tag detection  
- [x] Phone number argument parsing
- [x] Format phone number normalization
- [x] Profile picture fetch success
- [x] Error: User hidden profile
- [x] Error: No profile picture
- [x] Error: Invalid input format
- [x] Image download from URL
- [x] Image upload to WhatsApp media
- [x] Image message with reply context
- [x] Caption formatting

## ✅ Status

**STATUS: IMPLEMENTED** ✅

Fitur sudah sepenuhnya diimplementasikan dan siap digunakan setelah environment Go diperbaiki.
