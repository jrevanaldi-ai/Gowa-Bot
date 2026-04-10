# Gowa

[![Go Reference](https://pkg.go.dev/badge/github.com/jrevanaldi-ai/gowa.svg)](https://pkg.go.dev/github.com/jrevanaldi-ai/gowa)
[![Go Report Card](https://goreportcard.com/badge/github.com/jrevanaldi-ai/gowa)](https://goreportcard.com/report/github.com/jrevanaldi-ai/gowa)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL_2.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

**Gowa** adalah library Go lengkap untuk WhatsApp Web Multi-Device API dengan enkripsi end-to-end menggunakan Signal Protocol.

> 📦 **Module:** `github.com/jrevanaldi-ai/gowa`  
> 🐹 **Go Version:** 1.25.0+  
> 📜 **License:** Mozilla Public License v2.0  
> 👤 **Author:** Nathan ([@jrevanaldi-ai](https://github.com/jrevanaldi-ai))

---

## 📋 Daftar Isi

- [Fitur](#-fitur)
- [Instalasi](#-instalasi)
- [Quick Start](#-quick-start)
- [Dokumentasi Lengkap](#-dokumentasi-lengkap)
- [Struktur Project](#-struktur-project)
- [API Reference](#-api-reference)
- [Event System](#-event-system)
- [Encryption & Security](#-encryption--security)
- [Storage](#-storage)
- [Logging](#-logging)
- [Contoh Penggunaan](#-contoh-penggunaan)
- [Diskusi & Support](#-diskusi--support)
- [License](#-license)

---

## ✨ Fitur

### ✅ Fitur yang Sudah Diimplementasikan

#### 📨 Messaging
- ✅ Mengirim pesan teks ke chat pribadi dan grup
- ✅ Mengirim media (gambar, video, audio, dokumen, sticker)
- ✅ Menerima semua jenis pesan dengan dekripsi otomatis
- ✅ Reply, forward, dan quote messages
- ✅ Edit dan delete messages
- ✅ Reactions (emoji reactions)
- ✅ Polls (buat dan vote)
- ✅ Contact messages
- ✅ Location messages
- ✅ Bot messages (Meta AI integration)

#### 👥 Groups
- ✅ Membuat grup baru
- ✅ Mengelola participant (add, remove, promote, demote)
- ✅ Update info grup (nama, topik, foto, settings)
- ✅ Invite link generation dan management
- ✅ Join grup via invite link
- ✅ Get joined groups list
- ✅ Community management (link/unlink groups)
- ✅ Group announcement dan locked modes
- ✅ Ephemeral messages settings

#### 👤 User Features
- ✅ Cek nomor terdaftar di WhatsApp
- ✅ Get user info (avatar, status, devices)
- ✅ Business profile handling
- ✅ Profile picture management
- ✅ Push name updates
- ✅ Contact list sync
- ✅ Blocklist management

#### 📱 Multi-Device Support
- ✅ Pairing sebagai companion device
- ✅ QR code authentication
- ✅ Pairing code (8 karakter)
- ✅ Auto-reconnect dengan exponential backoff
- ✅ Session management
- ✅ Device list synchronization

#### 🔄 App State Sync
- ✅ Contact list synchronization
- ✅ Chat pin/mute/archive status
- ✅ Starred messages
- ✅ Labels dan labeling
- ✅ Settings sync (push name, locale)
- ✅ Protocol info sync
- ✅ 50+ index types support

#### 📊 Presence & Privacy
- ✅ Online/offline status
- ✅ Typing notifications (composing/paused)
- ✅ Recording audio status
- ✅ Privacy settings (last seen, profile, status, group add)
- ✅ Read receipts configuration
- ✅ Online presence settings

#### 📺 Newsletter (Channels)
- ✅ Get newsletter info
- ✅ List subscribed newsletters
- ✅ Create newsletter
- ✅ Follow/unfollow newsletter
- ✅ Mark newsletter as viewed
- ✅ Send reactions
- ✅ Get newsletter messages

#### 📢 Broadcast
- ✅ Status broadcast (experimental)
- ✅ Broadcast list recipients
- ✅ Status privacy settings

#### 📞 Calls
- ✅ Handle incoming call events
- ✅ Call offer/accept/terminate
- ✅ Call reject
- ⏳ Outgoing calls (coming soon)

#### 🔐 Security
- ✅ Signal Protocol encryption (libsignal)
- ✅ Noise_XX_25519_AESGCM_SHA256 handshake
- ✅ PreKey management
- ✅ SenderKey distribution (group messages)
- ✅ MessageSecret encryption
- ✅ Media encryption (AES-CBC, AES-GCM)
- ✅ HKDF key derivation
- ✅ LTHash for app state

#### 🛠️ Utilities
- ✅ Media upload/download dengan encryption
- ✅ Streaming download untuk file besar
- ✅ QR code channel dengan auto-rotation
- ✅ Event buffering
- ✅ Response waiter
- ✅ Message queue
- ✅ Version checking

---

## 📦 Instalasi

```bash
go get github.com/jrevanaldi-ai/gowa
```

### Dependencies

Gowa menggunakan dependencies berikut:

| Package | Version | Purpose |
|---------|---------|---------|
| `go.mau.fi/libsignal` | v0.2.1 | Signal Protocol encryption |
| `github.com/coder/websocket` | v1.8.14 | WebSocket client |
| `google.golang.org/protobuf` | v1.36.11 | Protocol Buffers |
| `golang.org/x/crypto` | v0.48.0 | Cryptographic functions |
| `github.com/rs/zerolog` | v1.34.0 | Logging |
| `github.com/google/uuid` | v1.6.0 | UUID generation |

---

## 🚀 Quick Start

### Contoh Dasar

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/jrevanaldi-ai/gowa"
    "github.com/jrevanaldi-ai/gowa/store"
    "github.com/jrevanaldi-ai/gowa/store/sqlstore"
    "github.com/jrevanaldi-ai/gowa/types/events"
    waLog "github.com/jrevanaldi-ai/gowa/util/log"
    "go.mau.fi/util/dbutil"
)

func main() {
    // Setup logging
    dbLog := waLog.Stdout("Database", "WARN", true)
    container, err := sqlstore.New(context.Background(), "sqlite3", "file:examplestore.db?_foreign_keys=on", dbLog)
    if err != nil {
        panic(err)
    }

    // Get or create device
    deviceStore, err := container.GetFirstDevice(context.Background())
    if err != nil {
        panic(err)
    }

    if deviceStore == nil || !deviceStore.ID.IsValid() {
        fmt.Println("No device found. Please pair first!")
        os.Exit(1)
    }

    // Setup client logging
    clientLog := waLog.Stdout("Client", "DEBUG", true)
    client := gowa.NewClient(deviceStore, clientLog)

    // Set event handler
    client.AddEventHandler(func(evt interface{}) {
        switch v := evt.(type) {
        case *events.Message:
            fmt.Printf("Message received: %s from %s\n", v.Message.GetConversation(), v.Info.SourceString())
        case *events.QR:
            fmt.Println("Scan QR code:")
            fmt.Println(v.Code)
        case *events.PairSuccess:
            fmt.Printf("Pairing successful! JID: %s\n", v.ID.String())
        case *events.Connected:
            fmt.Println("Connected to WhatsApp!")
        case *events.Disconnected:
            fmt.Println("Disconnected from WhatsApp")
        }
    })

    // Connect to WhatsApp servers
    err = client.Connect()
    if err != nil {
        panic(err)
    }

    // Wait for shutdown signal
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    <-c

    // Disconnect
    client.Disconnect()
}
```

### Pairing dengan QR Code

```go
// Setelah membuat client, start QR channel
qrChan, err := client.QRChannel(context.Background())
if err != nil {
    panic(err)
}

// Print QR codes
for qr := range qrChan {
    if qr.Event == gowa.QRChannelEventCode {
        fmt.Println("Scan this QR code:")
        fmt.Println(qr.Code)
    } else if qr.Event == gowa.QRChannelEventSuccess {
        fmt.Println("Pairing successful!")
        break
    } else if qr.Err != nil {
        panic(qr.Err)
    }
}
```

### Pairing dengan Code

```go
// Request pairing code
code, err := client.PairPhone("628123456789", true, "", "")
if err != nil {
    panic(err)
}
fmt.Printf("Enter this code on your phone: %s\n", code)
```

---

## 📚 Dokumentasi Lengkap

### Struktur Project

```
gowa/
├── appstate/              # App state processing
│   ├── lthash/           # LTHash implementation
│   ├── decode.go         # Decode app state patches
│   ├── encode.go         # Encode app state patches
│   ├── errors.go         # App state errors
│   ├── hash.go           # Hash state management
│   ├── keys.go           # App state key management
│   └── recovery.go       # App state recovery
├── binary/               # Binary XML encoding/decoding
│   ├── proto/            # Binary protobuf helpers
│   ├── token/            # Token definitions
│   ├── attrs.go          # Attribute utilities
│   ├── decoder.go        # Binary decoder
│   ├── encoder.go        # Binary encoder
│   ├── node.go           # XML Node structure
│   └── xml.go            # XML utilities
├── proto/                # Protocol Buffer definitions (40+ packages)
│   ├── waAdv/            # Device identity
│   ├── waE2E/            # End-to-End encrypted messages
│   ├── waCommon/         # Common types
│   ├── waGroupHistory/   # Group history
│   ├── waHistorySync/    # History synchronization
│   ├── waMsgApplication/ # Message application layer
│   ├── waMsgTransport/   # Message transport layer
│   ├── waMultiDevice/    # Multi-device support
│   └── ... (40+ packages)
├── socket/               # WebSocket & Noise protocol
│   ├── constants.go      # Socket constants
│   ├── framesocket.go    # Frame socket implementation
│   ├── noisehandshake.go # Noise handshake
│   └── noisesocket.go    # Noise socket implementation
├── store/                # Data storage interfaces
│   ├── sqlstore/         # SQL-backed storage
│   ├── clientpayload.go  # Client payload
│   ├── sessioncache.go   # Session cache
│   ├── signal.go         # Signal protocol store
│   └── store.go          # Store interfaces
├── types/                # Type definitions
│   ├── events/           # Event types
│   ├── botmap.go         # Bot JID mappings
│   ├── call.go           # Call types
│   ├── group.go          # Group types
│   ├── jid.go            # JID types
│   ├── message.go        # Message types
│   ├── newsletter.go     # Newsletter types
│   └── presence.go       # Presence types
├── util/                 # Utility packages
│   ├── cbcutil/          # CBC encryption
│   ├── gcmutil/          # GCM encryption
│   ├── hkdfutil/         # HKDF key derivation
│   ├── keys/             # Key management
│   └── log/              # Logging utilities
├── client.go             # Main client implementation
├── handshake.go          # Noise handshake
├── pair.go               # Pairing logic
├── message.go            # Message handling
├── send.go               # Message sending
├── group.go              # Group management
├── user.go               # User features
├── download.go           # Media download
├── upload.go             # Media upload
├── appstate.go           # App state sync
├── presence.go           # Presence handling
├── receipt.go            # Receipts handling
├── retry.go              # Retry logic
└── ... (40+ files)
```

---

## 🔌 API Reference

### Client Methods

#### Connection
| Method | Description |
|--------|-------------|
| `Connect()` | Connect ke WhatsApp servers |
| `Disconnect()` | Disconnect dari server |
| `IsConnected()` | Cek status koneksi |
| `IsLoggedIn()` | Cek apakah sudah login |

#### Pairing
| Method | Description |
|--------|-------------|
| `QRChannel(ctx)` | Get channel untuk QR codes |
| `PairPhone(phone, showPushNotification, deviceName, pairingCode)` | Pair dengan phone number |

#### Messaging
| Method | Description |
|--------|-------------|
| `SendMessage(chat, content, options)` | Kirim pesan ke chat/group |
| `SendFBMessage(chat, content, recipient, options)` | Kirim pesan format Facebook |
| `React(chat, messageID, reaction)` | Send reaction ke message |
| `SendPollCreation(chat, messageID, name, options, max)` | Buat poll |
| `SendPollVote(chat, messageID, vote)` | Vote poll |

#### Media
| Method | Description |
|--------|-------------|
| `Upload(mediaType, reader)` | Upload attachment |
| `Download(attachment)` | Download attachment dari message |
| `DownloadToFile(attachment, file)` | Download langsung ke file |

#### Groups
| Method | Description |
|--------|-------------|
| `CreateGroup(name, participants, options)` | Buat grup baru |
| `GetGroupInfo(group)` | Get info grup |
| `GetJoinedGroups()` | Get list semua grup |
| `UpdateGroupParticipants(group, participants, action)` | Add/remove/promote/demote |
| `SetGroupName(group, name)` | Update nama grup |
| `SetGroupTopic(group, oldID, newID, topic)` | Update topik grup |
| `SetGroupPhoto(group, jpeg)` | Update foto grup |
| `GetGroupInviteLink(group, reset)` | Get invite link |
| `JoinGroupWithLink(link)` | Join grup via link |
| `InviteIntoGroup(group, participants)` | Invite user ke grup |

#### User
| Method | Description |
|--------|-------------|
| `IsOnWhatsApp(phones)` | Cek nomor terdaftar |
| `GetUserInfo(jids)` | Get user info |
| `GetProfilePictureInfo(jid, options)` | Get profile picture URL |
| `GetBusinessProfile(jid)` | Get business profile |
| `GetBotListV2()` | Get bot list |
| `ResolveBusinessMessageLink(link)` | Resolve wa.me link |

#### App State
| Method | Description |
|--------|-------------|
| `FetchAppState(patchType, fullSync)` | Fetch patches dari server |
| `EncodePatch(patchType, mutations)` | Encode patch untuk dikirim |

#### Presence
| Method | Description |
|--------|-------------|
| `SendPresence(state)` | Set online/offline |
| `SubscribePresence(jid)` | Subscribe ke user presence |
| `SendChatPresence(chat, state, options)` | Typing notification |

#### Privacy
| Method | Description |
|--------|-------------|
| `GetPrivacySettings()` | Get current settings |
| `SetPrivacySetting(setting, value)` | Update privacy setting |

#### Newsletter
| Method | Description |
|--------|-------------|
| `GetNewsletterInfo(jid)` | Get newsletter metadata |
| `GetSubscribedNewsletters()` | List joined newsletters |
| `CreateNewsletter(name, options)` | Create newsletter |
| `FollowNewsletter(jid)` | Follow newsletter |
| `UnfollowNewsletter(jid)` | Unfollow newsletter |
| `GetNewsletterMessages(jid, count)` | Get messages |

#### Receipts
| Method | Description |
|--------|-------------|
| `MarkRead(ids, chat, sender)` | Mark messages as read |
| `SendMediaRetryReceipt(chat, messageID, mediaKey)` | Retry media download |

---

## 🎯 Event System

Gowa menggunakan event-driven architecture. Semua incoming data di-dispatch sebagai events.

### Connection Events

```go
type QR struct {
    Codes []string
}

type PairSuccess struct {
    ID  JID
    LID JID
}

type Connected struct{}

type Disconnected struct{}

type LoggedOut struct {
    Reason string
}
```

### Message Events

```go
type Message struct {
    Info    MessageInfo
    Message *waE2E.Message
    Data    *proto.Message
}

type Receipt struct {
    MessageSource
    MessageIDs []types.MessageID
    Type       types.ReceiptType
    Timestamp  time.Time
}

type UndecryptableMessage struct {
    Info MessageInfo
}
```

### Group Events

```go
type JoinedGroup struct {
    Reason string
    Create *waE2E.GroupCreateMessageContent
    Info   *types.GroupInfo
}

type GroupInfo struct {
    JID       JID
    Update    *types.GroupInfo
    RawUpdate *Node
}

type Picture struct {
    JID       JID
    Author    JID
    Type      string
    PictureID string
}
```

### User Events

```go
type Presence struct {
    JID      JID
    Available bool
}

type ChatPresence struct {
    JID      JID
    State    types.ChatPresence
    Media    types.ChatPresenceMedia
}

type PushName struct {
    JID      JID
    OldName  string
    NewName  string
}

type Contact struct {
    JID       JID
    FirstName string
    FullName  string
}
```

### App State Events

```go
type AppState struct {
    Index    string
    Value    *waSyncAction.SyncActionValue
    Action   *waSyncAction.SyncActionMessage
    SyncAction *waSyncAction.SyncActionMessage
}

type AppStateSyncComplete struct {
    Type appstate.WAPatchName
}
```

### Other Events

```go
type HistorySync struct {
    Data *waHistorySync.HistorySync
}

type MediaRetry struct {
    JID       JID
    Sender    JID
    MessageID MessageID
    Error     *MediaRetryError
    MediaKey  []byte
}

type NewsletterJoin struct {
    JID JID
}

type NewsletterLeave struct {
    JID JID
}

type CallOffer struct {
    CallID string
    From   JID
    // ... call details
}
```

### Event Handler Example

```go
client.AddEventHandler(func(evt interface{}) {
    switch v := evt.(type) {
    case *events.Message:
        if v.Info.IsFromMe {
            return
        }
        
        if v.Message.GetConversation() != "" {
            fmt.Printf("Text message: %s\n", v.Message.GetConversation())
        } else if v.Message.ImageMessage != nil {
            fmt.Println("Image message received")
            // Download image
            img, err := client.Download(v.Message.ImageMessage)
            if err != nil {
                log.Error("Failed to download image", "error", err)
            } else {
                // Process image
            }
        }
        
    case *events.Receipt:
        if v.Type == types.ReceiptTypeRead {
            fmt.Printf("Message %s read by %s\n", v.MessageIDs[0], v.Sender.String())
        }
        
    case *events.ChatPresence:
        if v.State == types.ChatPresenceComposing {
            fmt.Println("User is typing...")
        } else {
            fmt.Println("User stopped typing")
        }
        
    case *events.GroupInfo:
        fmt.Printf("Group info updated: %s\n", v.JID.String())
        
    case *events.Presence:
        status := "offline"
        if v.Available {
            status = "online"
        }
        fmt.Printf("User %s is %s\n", v.JID.String(), status)
    }
})
```

---

## 🔐 Encryption & Security

### Signal Protocol
Gowa menggunakan **libsignal** untuk enkripsi end-to-end:

- **PreKeys**: One-time keys untuk initial handshake
- **Sessions**: Stateful encryption untuk setiap contact
- **SenderKeys**: Efficient encryption untuk group messages
- **MessageSecret**: Additional encryption untuk reactions, polls, dll

### Noise Protocol
Handshake menggunakan **Noise_XX_25519_AESGCM_SHA256**:

1. **Client Hello**: Client mengirim ephemeral key
2. **Server Hello**: Server response dengan ephemeral + static key (encrypted)
3. **Certificate Verification**: Verifikasi sertifikat server
4. **Client Finish**: Client mengirim noise key + payload (encrypted)

### Media Encryption
Media dienkripsi dengan **AES-256-CBC** sebelum upload:

```go
// Generate media keys
mediaKey := hkdfutil.NewSHA256(mediaKeyRaw).Expand(nil, mediaType, 112)

// Split keys
encryptionKey := mediaKey[:32]
iv := mediaKey[32:48]
macKey := mediaKey[48:80]
```

### App State Hash
App state menggunakan **LTHash** (128-byte hash state) untuk integrity verification.

---

## 💾 Storage

Gowa menggunakan interface-based storage design yang pluggable.

### SQL Store
Support untuk berbagai database:

```go
// SQLite
container, _ := sqlstore.New(context.Background(), "sqlite3", "file:store.db?_foreign_keys=on", log)

// PostgreSQL
container, _ := sqlstore.New(context.Background(), "pgx", "postgres://user:pass@localhost/gowa", log)

// MySQL
container, _ := sqlstore.New(context.Background(), "mysql", "user:pass@tcp(localhost)/gowa", log)
```

### Store Interfaces

```go
type IdentityStore interface {
    GetIdentityKeyPair() (*keys.IdentityKeyPair, error)
    StoreIdentityKeyPair(*keys.IdentityKeyPair) error
    // ...
}

type SessionStore interface {
    HasSession(JID) (bool, error)
    GetSession(JID) (*state.Session, error)
    PutSession(JID, *state.Session) error
    // ...
}

type PreKeyStore interface {
    GetPreKey(uint32) (*keys.PreKey, error)
    MarkPreKeyUsed(uint32) error
    // ...
}

type SenderKeyStore interface {
    StoreSenderKey(JID, *state.SenderKey) error
    GetSenderKey(JID) (*state.SenderKey, error)
    // ...
}
```

### No-Op Store
Untuk testing atau custom implementation:

```go
device := store.NewDevice()
device.UseNoopStore()
```

---

## 📝 Logging

Gowa menggunakan **zerolog** untuk structured logging:

```go
// Setup logging
log := waLog.Stdout("Client", "DEBUG", true)
client := gowa.NewClient(deviceStore, log)

// Log levels: ERROR, WARN, INFO, DEBUG, TRACE
log := waLog.Stdout("Client", "WARN", true)

// Custom logger
type MyLogger struct{}
func (l MyLogger) Errorf(msg string, args ...interface{}) { log.Printf("ERROR: "+msg, args...) }
func (l MyLogger) Warnf(msg string, args ...interface{})  { log.Printf("WARN: "+msg, args...) }
func (l MyLogger) Infof(msg string, args ...interface{})  { log.Printf("INFO: "+msg, args...) }
func (l MyLogger) Debugf(msg string, args ...interface{}) { log.Printf("DEBUG: "+msg, args...) }
func (l MyLogger) Tracef(msg string, args ...interface{}) { log.Printf("TRACE: "+msg, args...) }

client := gowa.NewClient(deviceStore, MyLogger{})
```

---

## 💻 Contoh Penggunaan

### Bot Sederhana

```go
package main

import (
    "context"
    "strings"

    "github.com/jrevanaldi-ai/gowa"
    "github.com/jrevanaldi-ai/gowa/store/sqlstore"
    "github.com/jrevanaldi-ai/gowa/types/events"
    waLog "github.com/jrevanaldi-ai/gowa/util/log"
)

func main() {
    dbLog := waLog.Stdout("Database", "WARN", true)
    container, _ := sqlstore.New(context.Background(), "sqlite3", "file:bot.db?_foreign_keys=on", dbLog)
    deviceStore, _ := container.GetFirstDevice(context.Background())
    
    clientLog := waLog.Stdout("Client", "INFO", true)
    client := gowa.NewClient(deviceStore, clientLog)
    
    client.AddEventHandler(func(evt interface{}) {
        switch v := evt.(type) {
        case *events.Message:
            if v.Info.IsFromMe {
                return
            }
            
            msg := v.Message.GetConversation()
            if strings.HasPrefix(msg, "!ping") {
                client.SendMessage(context.Background(), v.Info.Chat, &gowa.Message{
                    Conversation: "Pong!",
                })
            } else if strings.HasPrefix(msg, "!echo ") {
                reply := strings.TrimPrefix(msg, "!echo ")
                client.SendMessage(context.Background(), v.Info.Chat, &gowa.Message{
                    Conversation: reply,
                })
            } else if strings.HasPrefix(msg, "!help") {
                help := `Commands:
!ping - Reply with pong
!echo <text> - Echo text
!help - Show this help`
                client.SendMessage(context.Background(), v.Info.Chat, &gowa.Message{
                    Conversation: help,
                })
            }
        }
    })
    
    client.Connect()
    select {}
}
```

### Media Downloader

```go
func downloadMedia(client *gowa.Client, msg *events.Message) {
    var mediaType string
    var attachment interface{}
    
    switch {
    case msg.Message.ImageMessage != nil:
        mediaType = "image"
        attachment = msg.Message.ImageMessage
    case msg.Message.VideoMessage != nil:
        mediaType = "video"
        attachment = msg.Message.VideoMessage
    case msg.Message.AudioMessage != nil:
        mediaType = "audio"
        attachment = msg.Message.AudioMessage
    case msg.Message.DocumentMessage != nil:
        mediaType = "document"
        attachment = msg.Message.DocumentMessage
    default:
        return
    }
    
    data, err := client.Download(attachment)
    if err != nil {
        log.Error("Download failed", "error", err)
        return
    }
    
    // Save to file
    filename := fmt.Sprintf("%s_%s.%s", mediaType, msg.Info.ID, getExtension(mediaType))
    err = os.WriteFile(filename, data, 0644)
    if err != nil {
        log.Error("Save failed", "error", err)
        return
    }
    
    log.Info("Downloaded", "file", filename)
}
```

### Group Manager

```go
func manageGroup(client *gowa.Client) {
    // Create group
    participants := []types.JID{
        types.NewJID("628123456789", types.DefaultUserServer),
        types.NewJID("628987654321", types.DefaultUserServer),
    }
    
    groupInfo, err := client.CreateGroup("My Group", participants, &gowa.ReqCreateGroup{
        Topic: "Group topic here",
    })
    if err != nil {
        log.Error("Create group failed", "error", err)
        return
    }
    
    log.Info("Group created", "jid", groupInfo.JID.String())
    
    // Update group name
    err = client.SetGroupName(groupInfo.JID, "New Group Name")
    if err != nil {
        log.Error("Update name failed", "error", err)
    }
    
    // Get invite link
    link, err := client.GetGroupInviteLink(groupInfo.JID, false)
    if err != nil {
        log.Error("Get invite link failed", "error", err)
    } else {
        log.Info("Invite link", "link", link)
    }
    
    // Promote participant to admin
    err = client.UpdateGroupParticipants(groupInfo.JID, participants[:1], gowa.ParticipantChangePromote)
    if err != nil {
        log.Error("Promote failed", "error", err)
    }
}
```

### Auto-Responder dengan AI

```go
func aiResponder(client *gowa.Client) {
    client.AddEventHandler(func(evt interface{}) {
        switch v := evt.(type) {
        case *events.Message:
            if v.Info.IsFromMe || v.Message.GetConversation() == "" {
                return
            }
            
            // Call AI API (contoh)
            response, err := callAI(v.Message.GetConversation())
            if err != nil {
                log.Error("AI call failed", "error", err)
                return
            }
            
            // Send reply
            _, err = client.SendMessage(context.Background(), v.Info.Chat, &gowa.Message{
                Conversation: response,
            }, gowa.SendRequestExtra{
                ReplyTo: v.Info.ID,
            })
            if err != nil {
                log.Error("Send failed", "error", err)
            }
        }
    })
}

func callAI(prompt string) (string, error) {
    // Implementasi call ke AI API (OpenAI, Anthropic, dll)
    return "AI response", nil
}
```

---

## 🗣️ Diskusi & Support

### Komunitas
- **Matrix Room**: [#whatsmeow:maunium.net](https://matrix.to/#/#whatsmeow:maunium.net)
- **GitHub Discussions**: [WhatsApp Protocol Q&A](https://github.com/jrevanaldi-ai/gowa/discussions/categories/whatsapp-protocol-q-a)

### Dokumentasi
- **Go Reference**: [pkg.go.dev/github.com/jrevanaldi-ai/gowa](https://pkg.go.dev/github.com/jrevanaldi-ai/gowa)
- **Examples**: Lihat folder `example/` untuk contoh lengkap

### Issues & Bugs
Laporkan bug atau request fitur di [GitHub Issues](https://github.com/jrevanaldi-ai/gowa/issues)

---

## ⚠️ Disclaimer

Library ini disediakan **AS IS** tanpa jaminan apapun. Penggunaan library ini adalah tanggung jawab pengguna. Pastikan untuk:

1. Mematuhi [Terms of Service WhatsApp](https://www.whatsapp.com/legal/terms-of-service)
2. Tidak menggunakan untuk spam atau aktivitas berbahaya
3. Menghormati privasi pengguna lain
4. Menggunakan dengan bijak dan bertanggung jawab

---

## 📄 License

**Gowa** dilisensikan di bawah [Mozilla Public License v2.0](https://opensource.org/licenses/MPL-2.0).

```
Copyright (c) 2025 Nathan (https://github.com/jrevanaldi-ai)

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at http://mozilla.org/MPL/2.0/.
```

---

## 🙏 Acknowledgments

- **WhatsApp Inc.** untuk WhatsApp Web API
- **Signal Protocol** untuk enkripsi end-to-end
- **mautrix/whatsmeow** sebagai inspirasi dan base library
- **Kontributor** yang telah membantu development

---

<div align="center">

**Made with ❤️ by [@jrevanaldi-ai](https://github.com/jrevanaldi-ai)**

⭐ Star this repo jika Anda merasa project ini bermanfaat!

</div>
