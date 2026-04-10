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


type BotClient struct {
	Client                *gowa.Client
	Registry              *lib.CommandRegistry
	Logger                *helper.Logger
	Cache                 *helper.Cache
	EphemeralHelper       *helper.EphemeralHelper
	JadibotSessionManager *helper.JadibotSessionManager
	Owners                map[string]bool
	SelfMode              bool
	mu                    sync.RWMutex
}


type BotConfig struct {
	Owners                []string
	Prefix                string
	MaxWorkers            int
	EnableCache           bool
	SelfMode              bool
	JadibotSessionManager *helper.JadibotSessionManager
}


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


func (b *BotClient) GetSelfMode() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.SelfMode
}


func NewBotClient(registry *lib.CommandRegistry, config *BotConfig) *BotClient {
	owners := make(map[string]bool)
	for _, owner := range config.Owners {
		owners[owner] = true
	}

	botClient := &BotClient{
		Registry:              registry,
		Logger:                helper.NewLogger("BotClient"),
		Cache:                 helper.NewCache(),
		Owners:                owners,
		SelfMode:              config.SelfMode,
		JadibotSessionManager: config.JadibotSessionManager,
	}


	botClient.EphemeralHelper = helper.NewEphemeralHelper(nil, 5*time.Minute)

	return botClient
}


func (b *BotClient) SetClient(client *gowa.Client) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Client = client


	if b.EphemeralHelper != nil {
		b.EphemeralHelper.SetClient(client)
	}
}


func (b *BotClient) SendMessage(ctx context.Context, chat types.JID, message *waE2E.Message) (interface{}, error) {

	if b.EphemeralHelper != nil {
		wrappedMsg, err := b.EphemeralHelper.WrapMessageWithEphemeral(ctx, chat, message)
		if err != nil {
			b.Logger.Warning("Failed to wrap ephemeral message: %v", err)

			wrappedMsg = message
		}
		message = wrappedMsg
	}


	return b.Client.SendMessage(ctx, chat, message)
}


func (b *BotClient) HandleMessage(ctx context.Context, evt *events.Message) {

	go b.processMessage(ctx, evt)
}


func (b *BotClient) processMessage(ctx context.Context, evt *events.Message) {

	b.mu.RLock()
	selfMode := b.SelfMode
	b.mu.RUnlock()


	if evt.Info.IsFromMe && !selfMode {
		return
	}


	if evt.Info.IsFromMe && selfMode {

		if !b.isOwner(evt.Info.Sender) {
			b.Logger.Debug("Self mode: Ignoring message from non-owner self")
			return
		}
	}


	var msg string


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

		if evt.Message.EditedMessage.Message != nil {

			b.processMessage(ctx, &events.Message{
				Info: evt.Info,
				Message: evt.Message.EditedMessage.Message,
			})
			return
		}
	case evt.Message.EphemeralMessage != nil:

		if evt.Message.EphemeralMessage.Message != nil {

			b.processMessage(ctx, &events.Message{
				Info: evt.Info,
				Message: evt.Message.EphemeralMessage.Message,
			})
			return
		}
	case evt.Message.ViewOnceMessage != nil:

		if evt.Message.ViewOnceMessage.Message != nil {
			b.processMessage(ctx, &events.Message{
				Info: evt.Info,
				Message: evt.Message.ViewOnceMessage.Message,
			})
			return
		}
	case evt.Message.ViewOnceMessageV2 != nil:

		if evt.Message.ViewOnceMessageV2.Message != nil {
			b.processMessage(ctx, &events.Message{
				Info: evt.Info,
				Message: evt.Message.ViewOnceMessageV2.Message,
			})
			return
		}
	case evt.Message.DocumentWithCaptionMessage != nil:

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


	isOwner := b.isOwner(evt.Info.Sender)


	if strings.HasPrefix(msg, "$") && isOwner {
		args := owner.ParseExecCommand(msg)
		if len(args) > 0 {

			b.handleExecCommand(ctx, evt, args)
			return
		}
	}


	cmd, args := b.parseCommandWithOwner(msg, isOwner)
	if cmd == "" {
		return
	}


	meta, found := b.Registry.GetCommand(cmd)
	if !found {
		return
	}


	if meta.OwnerOnly && !isOwner {
		return
	}


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


	handler, ok := b.Registry.GetHandler(cmd)
	if !ok {
		return
	}


	cmdCtx := &lib.CommandContext{
		Ctx:                     context.WithValue(ctx, "registry", b.Registry),
		Client:                  b.Client,
		BotClient:               b,
		JadibotSessionManager:   b.JadibotSessionManager,
		Sender:                  evt.Info.Sender,
		Chat:                    evt.Info.Chat,
		PushName:                evt.Info.PushName,
		IsGroup:                 evt.Info.IsGroup,
		IsOwner:                 isOwner,
		Message:                 msg,
		Args:                    args,
		MessageID:               evt.Info.ID,
		EphemeralWrapper: func(ctx context.Context, jid types.JID, message *waE2E.Message) (*waE2E.Message, error) {
			if b.EphemeralHelper != nil {
				return b.EphemeralHelper.WrapMessageWithEphemeral(ctx, jid, message)
			}
			return message, nil
		},
	}


	b.mu.RLock()
	client := b.Client
	b.mu.RUnlock()

	if client == nil {
		b.Logger.Error("Client is nil, cannot execute command")
		return
	}


	if err := handler(cmdCtx); err != nil {
		b.Logger.Error("Command error: %v", err)

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


func (b *BotClient) handleExecCommand(ctx context.Context, evt *events.Message, args []string) {

	cmdCtx := &lib.CommandContext{
		Ctx:                     context.WithValue(ctx, "registry", b.Registry),
		Client:                  b.Client,
		BotClient:               b,
		JadibotSessionManager:   b.JadibotSessionManager,
		Sender:                  evt.Info.Sender,
		Chat:                    evt.Info.Chat,
		PushName:                evt.Info.PushName,
		IsGroup:                 evt.Info.IsGroup,
		IsOwner:                 true,
		Message:                 evt.Message.GetConversation(),
		Args:                    args,
		MessageID:               evt.Info.ID,
		EphemeralWrapper: func(ctx context.Context, jid types.JID, message *waE2E.Message) (*waE2E.Message, error) {
			if b.EphemeralHelper != nil {
				return b.EphemeralHelper.WrapMessageWithEphemeral(ctx, jid, message)
			}
			return message, nil
		},
	}


	handler, ok := b.Registry.GetHandler("exec")
	if !ok {
		return
	}


	if err := handler(cmdCtx); err != nil {
		b.Logger.Error("Exec command error: %v", err)
	}
}


func (b *BotClient) parseCommandWithOwner(msg string, isOwner bool) (string, []string) {

	if isOwner {

		if strings.HasPrefix(msg, ".") {

			msg = strings.TrimPrefix(msg, ".")


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

			return "", nil
		} else {

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


	prefix := "."
	if !strings.HasPrefix(msg, prefix) {
		return "", nil
	}


	msg = strings.TrimPrefix(msg, prefix)


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


func (b *BotClient) parseCommand(msg string) (string, []string) {

	prefix := "."
	if strings.HasPrefix(msg, "$") {
		prefix = "$"
	}


	if !strings.HasPrefix(msg, prefix) {
		return "", nil
	}


	msg = strings.TrimPrefix(msg, prefix)


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


func (b *BotClient) isOwner(jid types.JID) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()


	if b.Owners[jid.String()] {
		return true
	}


	if b.Owners[jid.User] {
		return true
	}

	return false
}


func (b *BotClient) AddOwner(jid string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Owners[jid] = true
}


func (b *BotClient) RemoveOwner(jid string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.Owners, jid)
}


func (b *BotClient) EventHandler(evt any) {
	switch v := evt.(type) {
	case *events.Message:

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
