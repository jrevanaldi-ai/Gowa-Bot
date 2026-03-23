package helper

import (
	"context"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/jrevanaldi-ai/gowa"
	"github.com/jrevanaldi-ai/gowa/proto/waE2E"
	"github.com/jrevanaldi-ai/gowa/types"
)

// GroupConfig menyimpan konfigurasi untuk sebuah group
type GroupConfig struct {
	IsGroup           bool
	IsEphemeral       bool
	DisappearingTimer uint32
	CachedAt          time.Time
}

// EphemeralHelper helper untuk mendeteksi dan mengirim ephemeral message
type EphemeralHelper struct {
	client      *gowa.Client
	cache       map[types.JID]*GroupConfig
	cacheExpiry time.Duration
	mu          sync.RWMutex
	Logger      *Logger
}

// NewEphemeralHelper membuat EphemeralHelper baru
func NewEphemeralHelper(client *gowa.Client, cacheExpiry time.Duration) *EphemeralHelper {
	return &EphemeralHelper{
		client:      client,
		cache:       make(map[types.JID]*GroupConfig),
		cacheExpiry: cacheExpiry,
		Logger:      NewLogger("Ephemeral"),
	}
}

// SetClient mengatur client
func (h *EphemeralHelper) SetClient(client *gowa.Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.client = client
}

// GetGroupConfig mendapatkan konfigurasi group dengan cache
func (h *EphemeralHelper) GetGroupConfig(ctx context.Context, jid types.JID) (*GroupConfig, error) {
	h.mu.RLock()
	if config, ok := h.cache[jid]; ok {
		// Cek apakah cache masih valid
		if time.Since(config.CachedAt) < h.cacheExpiry {
			h.mu.RUnlock()
			return config, nil
		}
	}
	h.mu.RUnlock()

	// Cache miss atau expired, fetch dari server
	config, err := h.fetchGroupConfig(ctx, jid)
	if err != nil {
		return nil, err
	}

	// Simpan ke cache
	h.mu.Lock()
	config.CachedAt = time.Now()
	h.cache[jid] = config
	h.mu.Unlock()

	return config, nil
}

// fetchGroupConfig mengambil konfigurasi group dari server
func (h *EphemeralHelper) fetchGroupConfig(ctx context.Context, jid types.JID) (*GroupConfig, error) {
	config := &GroupConfig{
		IsGroup: jid.Server == types.GroupServer,
	}

	// Jika bukan group, return default
	if !config.IsGroup {
		return config, nil
	}

	h.Logger.Debug("Fetching group info for %s", jid.String())

	// Get group info dari server
	groupInfo, err := h.client.GetGroupInfo(ctx, jid)
	if err != nil {
		h.Logger.Warning("Failed to get group info: %v", err)
		// Jika error, return default config
		return config, nil
	}

	// Set ephemeral info
	config.IsEphemeral = groupInfo.IsEphemeral
	config.DisappearingTimer = groupInfo.DisappearingTimer

	h.Logger.Info("Group %s - Ephemeral: %v, Timer: %d seconds", 
		jid.String(), config.IsEphemeral, config.DisappearingTimer)

	return config, nil
}

// ClearCache membersihkan cache
func (h *EphemeralHelper) ClearCache() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cache = make(map[types.JID]*GroupConfig)
}

// RemoveCache menghapus cache untuk jid tertentu
func (h *EphemeralHelper) RemoveCache(jid types.JID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.cache, jid)
}

// WrapMessageWithEphemeral membungkus pesan dengan ephemeral jika group mendukung
func (h *EphemeralHelper) WrapMessageWithEphemeral(ctx context.Context, jid types.JID, message *waE2E.Message) (*waE2E.Message, error) {
	// Hanya proses di group
	if jid.Server != types.GroupServer {
		return message, nil
	}

	config, err := h.GetGroupConfig(ctx, jid)
	if err != nil {
		h.Logger.Warning("Failed to get group config: %v", err)
		return message, nil
	}

	h.Logger.Debug("Group %s - IsEphemeral: %v, Timer: %d", jid.String(), config.IsEphemeral, config.DisappearingTimer)

	// Jika group tidak ephemeral, return pesan asli
	if !config.IsEphemeral || config.DisappearingTimer == 0 {
		h.Logger.Debug("Ephemeral disabled for this group, sending normal message")
		return message, nil
	}

	// Bungkus pesan dengan ContextInfo yang memiliki Expiration timer
	h.Logger.Info("Wrapping message with ephemeral timer: %d seconds", config.DisappearingTimer)
	
	// Set expiration di ContextInfo
	contextInfo := &waE2E.ContextInfo{
		Expiration:              &config.DisappearingTimer,
		EphemeralSettingTimestamp: proto.Int64(time.Now().UnixMilli()),
	}
	
	// Wrap pesan berdasarkan tipe
	wrappedMessage := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			ContextInfo: contextInfo,
		},
	}
	
	// Copy isi pesan asli ke wrapped message
	if message.Conversation != nil {
		wrappedMessage.Conversation = message.Conversation
		wrappedMessage.ExtendedTextMessage.Text = message.Conversation
	} else if message.ExtendedTextMessage != nil {
		wrappedMessage.ExtendedTextMessage.Text = message.ExtendedTextMessage.Text
		// Copy ContextInfo dari pesan asli jika ada
		if message.ExtendedTextMessage.ContextInfo != nil {
			// Merge context info
			if message.ExtendedTextMessage.ContextInfo.StanzaID != nil {
				wrappedMessage.ExtendedTextMessage.ContextInfo.StanzaID = message.ExtendedTextMessage.ContextInfo.StanzaID
			}
			if message.ExtendedTextMessage.ContextInfo.Participant != nil {
				wrappedMessage.ExtendedTextMessage.ContextInfo.Participant = message.ExtendedTextMessage.ContextInfo.Participant
			}
			if message.ExtendedTextMessage.ContextInfo.ExternalAdReply != nil {
				wrappedMessage.ExtendedTextMessage.ContextInfo.ExternalAdReply = message.ExtendedTextMessage.ContextInfo.ExternalAdReply
			}
		}
	} else {
		// Untuk tipe pesan lain, gunakan EphemeralMessage wrapper
		return &waE2E.Message{
			EphemeralMessage: &waE2E.FutureProofMessage{
				Message: message,
			},
		}, nil
	}
	
	return wrappedMessage, nil
}

// SendMessageWithEphemeral mengirim pesan dengan dukungan ephemeral otomatis
func (h *EphemeralHelper) SendMessageWithEphemeral(ctx context.Context, jid types.JID, message *waE2E.Message, extra ...interface{}) (interface{}, error) {
	// Wrap message dengan ephemeral jika perlu
	wrappedMsg, err := h.WrapMessageWithEphemeral(ctx, jid, message)
	if err != nil {
		return nil, err
	}

	// Kirim pesan
	return h.client.SendMessage(ctx, jid, wrappedMsg)
}
