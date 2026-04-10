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


type GroupConfig struct {
	IsGroup           bool
	IsEphemeral       bool
	DisappearingTimer uint32
	CachedAt          time.Time
}


type EphemeralHelper struct {
	client      *gowa.Client
	cache       map[types.JID]*GroupConfig
	cacheExpiry time.Duration
	mu          sync.RWMutex
	Logger      *Logger
}


func NewEphemeralHelper(client *gowa.Client, cacheExpiry time.Duration) *EphemeralHelper {
	return &EphemeralHelper{
		client:      client,
		cache:       make(map[types.JID]*GroupConfig),
		cacheExpiry: cacheExpiry,
		Logger:      NewLogger("Ephemeral"),
	}
}


func (h *EphemeralHelper) SetClient(client *gowa.Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.client = client
}


func (h *EphemeralHelper) GetGroupConfig(ctx context.Context, jid types.JID) (*GroupConfig, error) {
	h.mu.RLock()
	if config, ok := h.cache[jid]; ok {

		if time.Since(config.CachedAt) < h.cacheExpiry {
			h.mu.RUnlock()
			return config, nil
		}
	}
	h.mu.RUnlock()


	config, err := h.fetchGroupConfig(ctx, jid)
	if err != nil {
		return nil, err
	}


	h.mu.Lock()
	config.CachedAt = time.Now()
	h.cache[jid] = config
	h.mu.Unlock()

	return config, nil
}


func (h *EphemeralHelper) fetchGroupConfig(ctx context.Context, jid types.JID) (*GroupConfig, error) {
	config := &GroupConfig{
		IsGroup: jid.Server == types.GroupServer,
	}


	if !config.IsGroup {
		return config, nil
	}

	h.Logger.Debug("Fetching group info for %s", jid.String())


	groupInfo, err := h.client.GetGroupInfo(ctx, jid)
	if err != nil {
		h.Logger.Warning("Failed to get group info: %v", err)

		return config, nil
	}


	config.IsEphemeral = groupInfo.IsEphemeral
	config.DisappearingTimer = groupInfo.DisappearingTimer

	h.Logger.Info("Group %s - Ephemeral: %v, Timer: %d seconds",
		jid.String(), config.IsEphemeral, config.DisappearingTimer)

	return config, nil
}


func (h *EphemeralHelper) ClearCache() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cache = make(map[types.JID]*GroupConfig)
}


func (h *EphemeralHelper) RemoveCache(jid types.JID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.cache, jid)
}


func (h *EphemeralHelper) WrapMessageWithEphemeral(ctx context.Context, jid types.JID, message *waE2E.Message) (*waE2E.Message, error) {

	if jid.Server != types.GroupServer {
		return message, nil
	}

	config, err := h.GetGroupConfig(ctx, jid)
	if err != nil {
		h.Logger.Warning("Failed to get group config: %v", err)
		return message, nil
	}

	h.Logger.Debug("Group %s - IsEphemeral: %v, Timer: %d", jid.String(), config.IsEphemeral, config.DisappearingTimer)


	if !config.IsEphemeral || config.DisappearingTimer == 0 {
		h.Logger.Debug("Ephemeral disabled for this group, sending normal message")
		return message, nil
	}


	h.Logger.Info("Wrapping message with ephemeral timer: %d seconds", config.DisappearingTimer)


	contextInfo := &waE2E.ContextInfo{
		Expiration:              &config.DisappearingTimer,
		EphemeralSettingTimestamp: proto.Int64(time.Now().UnixMilli()),
	}


	wrappedMessage := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			ContextInfo: contextInfo,
		},
	}


	if message.Conversation != nil {
		wrappedMessage.Conversation = message.Conversation
		wrappedMessage.ExtendedTextMessage.Text = message.Conversation
	} else if message.ExtendedTextMessage != nil {
		wrappedMessage.ExtendedTextMessage.Text = message.ExtendedTextMessage.Text

		if message.ExtendedTextMessage.ContextInfo != nil {

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

		return &waE2E.Message{
			EphemeralMessage: &waE2E.FutureProofMessage{
				Message: message,
			},
		}, nil
	}

	return wrappedMessage, nil
}


func (h *EphemeralHelper) SendMessageWithEphemeral(ctx context.Context, jid types.JID, message *waE2E.Message, extra ...interface{}) (interface{}, error) {

	wrappedMsg, err := h.WrapMessageWithEphemeral(ctx, jid, message)
	if err != nil {
		return nil, err
	}


	return h.client.SendMessage(ctx, jid, wrappedMsg)
}
