package client

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/jrevanaldi-ai/gowa"
	"github.com/jrevanaldi-ai/gowa/proto/waE2E"
	"github.com/jrevanaldi-ai/gowa/types"
	"github.com/jrevanaldi-ai/gowa/types/events"
	"github.com/jrevanaldi-ai/gowa-bot/commands/owner"
	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

// BotClient adalah wrapper untuk gowa.Client dengan fitur bot
type BotClient struct {
	Client          *gowa.Client
	Registry        *lib.CommandRegistry
	Logger          *helper.Logger
	Cache           *helper.Cache
	EphemeralHelper *helper.EphemeralHelper
	Owners          map[string]bool
	SelfMode        bool // Jika true, bot bisa merespon pesan dari diri sendiri
	mu              sync.RWMutex
}

// BotConfig adalah konfigurasi untuk bot
type BotConfig struct {
	Owners       []string
	Prefix       string
	MaxWorkers   int
	EnableCache  bool
	SelfMode     bool // Jika true, bot bisa merespon pesan dari diri sendiri
}

// SetSelfMode mengatur self mode
func (b *BotClient) SetSelfMode(mode bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.SelfMode = mode
	
	if mode {
		b.Logger.Info("Self mode activated")
	} else {
		b.Logger.Info("Public mode activated")
	}
}

// GetSelfMode mendapatkan status self mode
func (b *BotClient) GetSelfMode() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.SelfMode
}

// NewBotClient membuat BotClient baru
func NewBotClient(registry *lib.CommandRegistry, config *BotConfig) *BotClient {
	owners := make(map[string]bool)
	for _, owner := range config.Owners {
		owners[owner] = true
	}

	botClient := &BotClient{
		Registry: registry,
		Logger:   helper.NewLogger("BotClient"),
		Cache:    helper.NewCache(),
		Owners:   owners,
		SelfMode: config.SelfMode,
	}

	// Init EphemeralHelper
	botClient.EphemeralHelper = helper.NewEphemeralHelper(nil, 5*time.Minute)

	return botClient
}

// SetClient mengatur gowa.Client
func (b *BotClient) SetClient(client *gowa.Client) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Client = client

	// Set client ke EphemeralHelper
	if b.EphemeralHelper != nil {
		b.EphemeralHelper.SetClient(client)
	}
}

// SendMessage mengirim pesan dengan dukungan ephemeral otomatis
func (b *BotClient) SendMessage(ctx context.Context, chat types.JID, message *waE2E.Message) (interface{}, error) {
	// Gunakan EphemeralHelper untuk wrap message jika perlu
	if b.EphemeralHelper != nil {
		wrappedMsg, err := b.EphemeralHelper.WrapMessageWithEphemeral(ctx, chat, message)
		if err != nil {
			b.Logger.Warning("Failed to wrap ephemeral message: %v", err)
			// Fallback ke pesan asli
			wrappedMsg = message
		}
		message = wrappedMsg
	}

	// Kirim pesan
	return b.Client.SendMessage(ctx, chat, message)
}

// HandleMessage menangani pesan masuk dengan goroutine
func (b *BotClient) HandleMessage(ctx context.Context, evt *events.Message) {
	// Gunakan goroutine untuk handle setiap pesan
	go b.processMessage(ctx, evt)
}

// processMessage memproses pesan masuk
func (b *BotClient) processMessage(ctx context.Context, evt *events.Message) {
	// Get self mode status
	b.mu.RLock()
	selfMode := b.SelfMode
	b.mu.RUnlock()

	// Ignore pesan dari diri sendiri jika bukan self mode
	if evt.Info.IsFromMe && !selfMode {
		return
	}

	// Jika self mode aktif dan pesan dari diri sendiri, pastikan hanya owner yang bisa pakai
	if evt.Info.IsFromMe && selfMode {
		// Cek apakah nomor bot adalah owner
		if !b.isOwner(evt.Info.Sender) {
			b.Logger.Debug("Self mode: Ignoring message from non-owner self")
			return
		}
	}

	// Get text dari berbagai tipe pesan
	var msg string
	
	// Cek semua tipe pesan yang mungkin mengandung text
	switch {
	case evt.Message.Conversation != nil:
		msg = *evt.Message.Conversation
	case evt.Message.ExtendedTextMessage != nil && evt.Message.ExtendedTextMessage.Text != nil:
		msg = *evt.Message.ExtendedTextMessage.Text
	case evt.Message.ImageMessage != nil && evt.Message.ImageMessage.Caption != nil:
		msg = *evt.Message.ImageMessage.Caption
	case evt.Message.VideoMessage != nil && evt.Message.VideoMessage.Caption != nil:
		msg = *evt.Message.VideoMessage.Caption
	case evt.Message.DocumentMessage != nil && evt.Message.DocumentMessage.Caption != nil:
		msg = *evt.Message.DocumentMessage.Caption
	case evt.Message.LocationMessage != nil && evt.Message.LocationMessage.Comment != nil:
		msg = *evt.Message.LocationMessage.Comment
	case evt.Message.LiveLocationMessage != nil && evt.Message.LiveLocationMessage.Caption != nil:
		msg = *evt.Message.LiveLocationMessage.Caption
	case evt.Message.StickerMessage != nil:
		msg = "Sticker"
	case evt.Message.ContactMessage != nil && evt.Message.ContactMessage.DisplayName != nil:
		msg = *evt.Message.ContactMessage.DisplayName
	case evt.Message.ButtonsResponseMessage != nil && evt.Message.ButtonsResponseMessage.SelectedButtonID != nil:
		msg = *evt.Message.ButtonsResponseMessage.SelectedButtonID
	case evt.Message.ListResponseMessage != nil && evt.Message.ListResponseMessage.SingleSelectReply != nil:
		if evt.Message.ListResponseMessage.SingleSelectReply.SelectedRowID != nil {
			msg = *evt.Message.ListResponseMessage.SingleSelectReply.SelectedRowID
		}
	case evt.Message.InteractiveResponseMessage != nil && evt.Message.InteractiveResponseMessage.Body != nil:
		if evt.Message.InteractiveResponseMessage.Body.Text != nil {
			msg = *evt.Message.InteractiveResponseMessage.Body.Text
		}
	case evt.Message.ReactionMessage != nil && evt.Message.ReactionMessage.Text != nil:
		msg = *evt.Message.ReactionMessage.Text
	case evt.Message.PollCreationMessage != nil && evt.Message.PollCreationMessage.Name != nil:
		msg = *evt.Message.PollCreationMessage.Name
	case evt.Message.PollUpdateMessage != nil:
		msg = "Poll Vote"
	case evt.Message.OrderMessage != nil && evt.Message.OrderMessage.OrderTitle != nil:
		msg = *evt.Message.OrderMessage.OrderTitle
	case evt.Message.RequestPhoneNumberMessage != nil:
		msg = "Request Phone Number"
	case evt.Message.CallLogMesssage != nil:
		msg = "Call Log"
	case evt.Message.ScheduledCallCreationMessage != nil:
		msg = "Scheduled Call"
	case evt.Message.GroupInviteMessage != nil:
		msg = "Group Invite"
	case evt.Message.TemplateButtonReplyMessage != nil && evt.Message.TemplateButtonReplyMessage.SelectedID != nil:
		msg = *evt.Message.TemplateButtonReplyMessage.SelectedID
	case evt.Message.ProductMessage != nil && evt.Message.ProductMessage.Product != nil:
		if evt.Message.ProductMessage.Product.Title != nil {
			msg = *evt.Message.ProductMessage.Product.Title
		}
	case evt.Message.ListMessage != nil && evt.Message.ListMessage.Title != nil:
		msg = *evt.Message.ListMessage.Title
	case evt.Message.EditedMessage != nil:
		// Handle edited message - recurse
		if evt.Message.EditedMessage.Message != nil {
			// Process edited message
			b.processMessage(ctx, &events.Message{
				Info: evt.Info,
				Message: evt.Message.EditedMessage.Message,
			})
			return
		}
	case evt.Message.EphemeralMessage != nil:
		// Handle ephemeral message - recurse
		if evt.Message.EphemeralMessage.Message != nil {
			// Process ephemeral message
			b.processMessage(ctx, &events.Message{
				Info: evt.Info,
				Message: evt.Message.EphemeralMessage.Message,
			})
			return
		}
	case evt.Message.ViewOnceMessage != nil:
		// Handle view once message - recurse
		if evt.Message.ViewOnceMessage.Message != nil {
			b.processMessage(ctx, &events.Message{
				Info: evt.Info,
				Message: evt.Message.ViewOnceMessage.Message,
			})
			return
		}
	case evt.Message.ViewOnceMessageV2 != nil:
		// Handle view once v2 - recurse
		if evt.Message.ViewOnceMessageV2.Message != nil {
			b.processMessage(ctx, &events.Message{
				Info: evt.Info,
				Message: evt.Message.ViewOnceMessageV2.Message,
			})
			return
		}
	case evt.Message.DocumentWithCaptionMessage != nil:
		// Handle document with caption - recurse
		if evt.Message.DocumentWithCaptionMessage.Message != nil {
			b.processMessage(ctx, &events.Message{
				Info: evt.Info,
				Message: evt.Message.DocumentWithCaptionMessage.Message,
			})
			return
		}
	}

	if msg == "" {
		return
	}

	// Cek apakah sender adalah owner
	isOwner := b.isOwner(evt.Info.Sender)

	// Handle exec command dengan prefix $ terlebih dahulu
	if strings.HasPrefix(msg, "$") && isOwner {
		args := owner.ParseExecCommand(msg)
		if len(args) > 0 {
			// Langsung handle exec command
			b.handleExecCommand(ctx, evt, args)
			return
		}
	}

	// Parse command dengan atau tanpa prefix
	cmd, args := b.parseCommandWithOwner(msg, isOwner)
	if cmd == "" {
		return
	}

	// Dapatkan metadata command
	meta, found := b.Registry.GetCommand(cmd)
	if !found {
		return
	}

	// Cek owner only
	if meta.OwnerOnly && !isOwner {
		return
	}

	// Log pesan masuk
	chatType := "Private"
	if evt.Info.IsGroup {
		chatType = "Group"
	}
	b.Logger.Message(
		evt.Info.PushName,
		evt.Info.Sender.String(),
		cmd,
		chatType,
	)

	// Dapatkan handler
	handler, ok := b.Registry.GetHandler(cmd)
	if !ok {
		return
	}

	// Buat command context
	cmdCtx := &lib.CommandContext{
		Ctx:         context.WithValue(ctx, "registry", b.Registry),
		Client:      b.Client,
		BotClient:   b, // Set BotClient reference
		Sender:      evt.Info.Sender,
		Chat:        evt.Info.Chat,
		PushName:    evt.Info.PushName,
		IsGroup:     evt.Info.IsGroup,
		IsOwner:     isOwner,
		Message:     msg,
		Args:        args,
		MessageID:   evt.Info.ID,
		EphemeralWrapper: func(ctx context.Context, jid types.JID, message *waE2E.Message) (*waE2E.Message, error) {
			if b.EphemeralHelper != nil {
				return b.EphemeralHelper.WrapMessageWithEphemeral(ctx, jid, message)
			}
			return message, nil
		},
	}

	// Eksekusi handler dengan error handling
	b.mu.RLock()
	client := b.Client
	b.mu.RUnlock()

	if client == nil {
		b.Logger.Error("Client is nil, cannot execute command")
		return
	}

	// Handle error dari handler
	if err := handler(cmdCtx); err != nil {
		b.Logger.Error("Command error: %v", err)
		// Kirim error message ke user
		errorMsg := fmt.Sprintf("❌ Terjadi kesalahan: %v", err)
		_, _ = b.SendMessage(ctx, evt.Info.Chat, &waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text: &errorMsg,
				ContextInfo: &waE2E.ContextInfo{
					StanzaID:    proto.String(evt.Info.ID),
					Participant: proto.String(evt.Info.Sender.String()),
				},
			},
		})
	}
}

// handleExecCommand menangani exec command secara langsung
func (b *BotClient) handleExecCommand(ctx context.Context, evt *events.Message, args []string) {
	// Buat command context
	cmdCtx := &lib.CommandContext{
		Ctx:         context.WithValue(ctx, "registry", b.Registry),
		Client:      b.Client,
		Sender:      evt.Info.Sender,
		Chat:        evt.Info.Chat,
		PushName:    evt.Info.PushName,
		IsGroup:     evt.Info.IsGroup,
		IsOwner:     true,
		Message:     evt.Message.GetConversation(),
		Args:        args,
		MessageID:   evt.Info.ID,
		EphemeralWrapper: func(ctx context.Context, jid types.JID, message *waE2E.Message) (*waE2E.Message, error) {
			if b.EphemeralHelper != nil {
				return b.EphemeralHelper.WrapMessageWithEphemeral(ctx, jid, message)
			}
			return message, nil
		},
	}

	// Dapatkan handler exec
	handler, ok := b.Registry.GetHandler("exec")
	if !ok {
		return
	}

	// Eksekusi handler
	if err := handler(cmdCtx); err != nil {
		b.Logger.Error("Exec command error: %v", err)
	}
}

// parseCommandWithOwner memparse command dengan atau tanpa prefix untuk owner
func (b *BotClient) parseCommandWithOwner(msg string, isOwner bool) (string, []string) {
	// Jika owner, bisa pakai tanpa prefix
	if isOwner {
		// Cek apakah pesan dimulai dengan prefix "." atau "$"
		if strings.HasPrefix(msg, ".") {
			// Hapus prefix
			msg = strings.TrimPrefix(msg, ".")
			
			// Split command dan args
			parts := strings.Fields(msg)
			if len(parts) == 0 {
				return "", nil
			}
			
			cmd := strings.ToLower(parts[0])
			var args []string
			if len(parts) > 1 {
				args = parts[1:]
			}
			return cmd, args
		} else if strings.HasPrefix(msg, "$") {
			// Exec command - akan dihandle terpisah
			return "", nil
		} else {
			// Tanpa prefix untuk owner
			parts := strings.Fields(msg)
			if len(parts) == 0 {
				return "", nil
			}
			
			cmd := strings.ToLower(parts[0])
			var args []string
			if len(parts) > 1 {
				args = parts[1:]
			}
			return cmd, args
		}
	}
	
	// Untuk non-owner, harus pakai prefix "."
	prefix := "."
	if !strings.HasPrefix(msg, prefix) {
		return "", nil
	}
	
	// Hapus prefix
	msg = strings.TrimPrefix(msg, prefix)
	
	// Split command dan args
	parts := strings.Fields(msg)
	if len(parts) == 0 {
		return "", nil
	}
	
	cmd := strings.ToLower(parts[0])
	var args []string
	if len(parts) > 1 {
		args = parts[1:]
	}
	
	return cmd, args
}

// parseCommand memparse command dari pesan
func (b *BotClient) parseCommand(msg string) (string, []string) {
	// Cek prefix "." atau "$"
	prefix := "."
	if strings.HasPrefix(msg, "$") {
		prefix = "$"
	}

	// Cek apakah pesan dimulai dengan prefix
	if !strings.HasPrefix(msg, prefix) {
		return "", nil
	}

	// Hapus prefix
	msg = strings.TrimPrefix(msg, prefix)

	// Split command dan args
	parts := strings.Fields(msg)
	if len(parts) == 0 {
		return "", nil
	}

	cmd := strings.ToLower(parts[0])
	var args []string
	if len(parts) > 1 {
		args = parts[1:]
	}

	return cmd, args
}

// isOwner cek apakah user adalah owner
func (b *BotClient) isOwner(jid types.JID) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Cek dengan JID string
	if b.Owners[jid.String()] {
		return true
	}

	// Cek dengan user saja (tanpa server)
	if b.Owners[jid.User] {
		return true
	}

	return false
}

// AddOwner menambahkan owner baru
func (b *BotClient) AddOwner(jid string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Owners[jid] = true
}

// RemoveOwner menghapus owner
func (b *BotClient) RemoveOwner(jid string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.Owners, jid)
}

// Event Handler untuk WhatsApp events
func (b *BotClient) EventHandler(evt any) {
	switch v := evt.(type) {
	case *events.Message:
		// Handle pesan dengan context background
		ctx := context.Background()
		b.HandleMessage(ctx, v)

	case *events.Connected:
		b.Logger.Success("Connected to WhatsApp!")

	case *events.LoggedOut:
		b.Logger.Warning("Logged out from WhatsApp!")

	case *events.Disconnected:
		b.Logger.Warning("Disconnected from WhatsApp!")

	case *events.QR:
		b.Logger.Info("QR code available (use PairPhone instead)")

	case *events.PairSuccess:
		b.Logger.Success("Successfully paired with phone: %s", v.ID.String())

	case *events.PairError:
		b.Logger.Error("Pairing failed: %v", v.Error)
	}
}
