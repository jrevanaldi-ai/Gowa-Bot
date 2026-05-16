package client

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/jrevanaldi-ai/gowa"
	"github.com/jrevanaldi-ai/gowa-bot/commands/owner"
	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
	"github.com/jrevanaldi-ai/gowa/proto/waE2E"
	"github.com/jrevanaldi-ai/gowa/types"
	"github.com/jrevanaldi-ai/gowa/types/events"
)

type BotClient struct {
	Client                *gowa.Client
	Registry              *lib.CommandRegistry
	Logger                *helper.Logger
	Cache                 *helper.Cache
	EphemeralHelper       *helper.EphemeralHelper
	JadibotSessionManager *helper.JadibotSessionManager
	DBManager             *helper.DatabaseManager
	Dispatcher            *lib.Dispatcher
	Owners                map[string]bool
	SelfMode              bool
	IsMainBot             bool
	Prefixes              []string
	mu                    sync.RWMutex
}

type BotConfig struct {
	Owners                []string
	Prefixes              []string
	MaxWorkers            int
	EnableCache           bool
	SelfMode              bool
	IsMainBot             bool
	JadibotSessionManager *helper.JadibotSessionManager
	DBManager             *helper.DatabaseManager
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

func (b *BotClient) GetDBManager() interface{} {
	return b.DBManager
}

func (b *BotClient) GetCache() interface{} {
	return b.Cache
}

func (b *BotClient) SetPrefixes(prefixes []string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Prefixes = prefixes
	b.Logger.Info("Prefixes updated: %v", prefixes)
}

func (b *BotClient) GetPrefixes() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Prefixes
}

func NewBotClient(registry *lib.CommandRegistry, config *BotConfig) *BotClient {
	owners := make(map[string]bool)
	for _, owner := range config.Owners {
		owners[owner] = true
	}

	prefixes := config.Prefixes
	if len(prefixes) == 0 {
		prefixes = []string{"."}
	}

	maxWorkers := config.MaxWorkers
	if maxWorkers <= 0 {
		maxWorkers = 10
	}

	botClient := &BotClient{
		Registry:              registry,
		Logger:                helper.NewLogger("BotClient"),
		Cache:                 helper.NewCache(),
		Dispatcher:            lib.NewDispatcher(maxWorkers),
		Owners:                owners,
		SelfMode:              config.SelfMode,
		IsMainBot:             config.IsMainBot,
		JadibotSessionManager: config.JadibotSessionManager,
		DBManager:             config.DBManager,
		Prefixes:              prefixes,
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
	if b.Dispatcher != nil {
		b.Dispatcher.Run(func() { b.processMessage(ctx, evt) })
		return
	}
	go b.processMessage(ctx, evt)
}

func (b *BotClient) processMessage(ctx context.Context, evt *events.Message) {

	b.mu.RLock()
	selfMode := b.SelfMode
	b.mu.RUnlock()

	isOwner := b.isOwner(evt.Info.Sender)

	if evt.Info.IsFromMe && !selfMode && !isOwner {
		return
	}

	if evt.Info.IsFromMe && selfMode && !isOwner {
		return
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
				Info:    evt.Info,
				Message: evt.Message.EditedMessage.Message,
			})
			return
		}
	case evt.Message.EphemeralMessage != nil:

		if evt.Message.EphemeralMessage.Message != nil {

			b.processMessage(ctx, &events.Message{
				Info:    evt.Info,
				Message: evt.Message.EphemeralMessage.Message,
			})
			return
		}
	case evt.Message.ViewOnceMessage != nil:

		if evt.Message.ViewOnceMessage.Message != nil {
			b.processMessage(ctx, &events.Message{
				Info:    evt.Info,
				Message: evt.Message.ViewOnceMessage.Message,
			})
			return
		}
	case evt.Message.ViewOnceMessageV2 != nil:

		if evt.Message.ViewOnceMessageV2.Message != nil {
			b.processMessage(ctx, &events.Message{
				Info:    evt.Info,
				Message: evt.Message.ViewOnceMessageV2.Message,
			})
			return
		}
	case evt.Message.DocumentWithCaptionMessage != nil:

		if evt.Message.DocumentWithCaptionMessage.Message != nil {
			b.processMessage(ctx, &events.Message{
				Info:    evt.Info,
				Message: evt.Message.DocumentWithCaptionMessage.Message,
			})
			return
		}
	}

	if msg == "" {
		return
	}

	if b.DBManager != nil && !isOwner {

		if evt.Info.IsGroup {
			isGroupBanned, err := b.DBManager.IsBanned(evt.Info.Chat.String(), "group")
			if err != nil {
				b.Logger.Warning("Failed to check group ban status: %v", err)
			} else if isGroupBanned {
				b.Logger.Debug("Ignoring command from banned group: %s", evt.Info.Chat.String())
				return
			}
		}

		isUserBanned, err := b.DBManager.IsBanned(evt.Info.Sender.String(), "user")
		if err != nil {
			b.Logger.Warning("Failed to check user ban status: %v", err)
		} else if isUserBanned {
			b.Logger.Debug("Ignoring command from banned user: %s", evt.Info.Sender.String())
			return
		}
	}

	if strings.HasPrefix(msg, "$") && isOwner {
		args := owner.ParseExecCommand(msg)
		if len(args) > 0 {

			b.handleExecCommand(ctx, evt, args)
			return
		}
	}

	var replyMsg *lib.ReplyMessageInfo
	var mentions []string
	if evt.Message.ExtendedTextMessage != nil && evt.Message.ExtendedTextMessage.ContextInfo != nil {
		contextInfo := evt.Message.ExtendedTextMessage.ContextInfo
		if contextInfo.StanzaID != nil && contextInfo.Participant != nil {
			replyMsg = &lib.ReplyMessageInfo{
				MessageID: *contextInfo.StanzaID,
				Sender:    *contextInfo.Participant,
				Message:   helper.ExtractMessageText(contextInfo.QuotedMessage),
			}
		}
		if len(contextInfo.MentionedJID) > 0 {
			mentions = make([]string, len(contextInfo.MentionedJID))
			for i, mention := range contextInfo.MentionedJID {
				mentions[i] = mention
			}
		}
	}

	cmd, args := b.parseCommandWithOwner(msg, isOwner, evt.Info.IsFromMe)

	if cmd == "" || (cmd != "lune" && !b.Registry.IsCommand(cmd)) {
		isLuneKeyword := containsWord(msg, "lune")
		isAIReply := replyMsg != nil && helper.IsAIReply(b.Cache, evt.Info.Chat.String(), replyMsg.MessageID)

		if isLuneKeyword || isAIReply {
			if handler, ok := b.Registry.GetHandler("lune"); ok {

				chatType := "Private"
				if evt.Info.IsGroup {
					chatType = "Group"
				}
				trigger := "lune (keyword)"
				if isAIReply {
					trigger = "lune (reply)"
				}
				b.Logger.Message(evt.Info.PushName, evt.Info.Sender.User, chatType, trigger, extractMediaSize(evt.Message))

				cmdCtx := &lib.CommandContext{
					Ctx:                   context.WithValue(context.WithValue(ctx, "registry", b.Registry), "gowa_client", b.Client),
					Client:                b.Client,
					BotClient:             b,
					JadibotSessionManager: b.JadibotSessionManager,
					Sender:                evt.Info.Sender,
					Chat:                  evt.Info.Chat,
					PushName:              evt.Info.PushName,
					IsGroup:               evt.Info.IsGroup,
					IsOwner:               isOwner,
					Message:               msg,
					Args:                  strings.Fields(msg),
					MessageID:             evt.Info.ID,
					EphemeralWrapper: func(ctx context.Context, jid types.JID, message *waE2E.Message) (*waE2E.Message, error) {
						if b.EphemeralHelper != nil {
							return b.EphemeralHelper.WrapMessageWithEphemeral(ctx, jid, message)
						}
						return message, nil
					},
					ReplyMessage: replyMsg,
					Mentions:     mentions,
					RawEvent:     evt,
					RawMessage:   evt.Message,
				}

				if err := handler(cmdCtx); err != nil {
					b.Logger.Error("Keyword AI error: %v", err)
				}
				return
			}
		}
	}

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
		evt.Info.Sender.User,
		chatType,
		cmd,
		extractMediaSize(evt.Message),
	)

	handler, ok := b.Registry.GetHandler(cmd)
	if !ok {
		return
	}

	cmdCtx := &lib.CommandContext{
		Ctx:                   context.WithValue(context.WithValue(ctx, "registry", b.Registry), "gowa_client", b.Client),
		Client:                b.Client,
		BotClient:             b,
		JadibotSessionManager: b.JadibotSessionManager,
		Sender:                evt.Info.Sender,
		Chat:                  evt.Info.Chat,
		PushName:              evt.Info.PushName,
		IsGroup:               evt.Info.IsGroup,
		IsOwner:               isOwner,
		Message:               msg,
		Args:                  args,
		MessageID:             evt.Info.ID,
		EphemeralWrapper: func(ctx context.Context, jid types.JID, message *waE2E.Message) (*waE2E.Message, error) {
			if b.EphemeralHelper != nil {
				return b.EphemeralHelper.WrapMessageWithEphemeral(ctx, jid, message)
			}
			return message, nil
		},
		ReplyMessage: replyMsg,
		Mentions:     mentions,
		RawEvent:     evt,
		RawMessage:   evt.Message,
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

		errorMsg := fmt.Sprintf("Terjadi kesalahan: %v", err)
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
		Ctx:                   context.WithValue(ctx, "registry", b.Registry),
		Client:                b.Client,
		BotClient:             b,
		JadibotSessionManager: b.JadibotSessionManager,
		Sender:                evt.Info.Sender,
		Chat:                  evt.Info.Chat,
		PushName:              evt.Info.PushName,
		IsGroup:               evt.Info.IsGroup,
		IsOwner:               true,
		Message:               evt.Message.GetConversation(),
		Args:                  args,
		MessageID:             evt.Info.ID,
		EphemeralWrapper: func(ctx context.Context, jid types.JID, message *waE2E.Message) (*waE2E.Message, error) {
			if b.EphemeralHelper != nil {
				return b.EphemeralHelper.WrapMessageWithEphemeral(ctx, jid, message)
			}
			return message, nil
		},
		RawEvent:   evt,
		RawMessage: evt.Message,
	}

	handler, ok := b.Registry.GetHandler("exec")
	if !ok {
		return
	}

	if err := handler(cmdCtx); err != nil {
		b.Logger.Error("Exec command error: %v", err)
	}
}

func containsWord(msg, word string) bool {
	lower := strings.ToLower(msg)
	w := strings.ToLower(word)
	idx := 0
	for {
		found := strings.Index(lower[idx:], w)
		if found < 0 {
			return false
		}
		start := idx + found
		end := start + len(w)
		if (start == 0 || !isWordChar(lower[start-1])) && (end == len(lower) || !isWordChar(lower[end])) {
			return true
		}
		idx = start + 1
	}
}

func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_'
}

func extractMediaSize(m *waE2E.Message) string {
	if m == nil {
		return "-"
	}
	var size uint64
	switch {
	case m.ImageMessage != nil && m.ImageMessage.FileLength != nil:
		size = *m.ImageMessage.FileLength
	case m.VideoMessage != nil && m.VideoMessage.FileLength != nil:
		size = *m.VideoMessage.FileLength
	case m.AudioMessage != nil && m.AudioMessage.FileLength != nil:
		size = *m.AudioMessage.FileLength
	case m.DocumentMessage != nil && m.DocumentMessage.FileLength != nil:
		size = *m.DocumentMessage.FileLength
	case m.StickerMessage != nil && m.StickerMessage.FileLength != nil:
		size = *m.StickerMessage.FileLength
	default:
		return "-"
	}
	return helper.HumanSize(size)
}

func (b *BotClient) parseCommandWithOwner(msg string, isOwner bool, forcePrefix bool) (string, []string) {

	b.mu.RLock()
	prefixes := b.Prefixes
	b.mu.RUnlock()

	for _, prefix := range prefixes {
		if strings.HasPrefix(msg, prefix) {
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
	}

	if isOwner && !forcePrefix {

		if strings.HasPrefix(msg, "$") {
			return "", nil
		}

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

	return "", nil
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

	if b.Client != nil && b.Client.Store != nil && b.Client.Store.ID != nil {
		if jid.User == b.Client.Store.ID.User {
			return true
		}
	}

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
