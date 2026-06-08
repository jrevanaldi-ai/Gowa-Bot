package client

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	"go.mau.fi/whatsmeow"
	"github.com/jrevanaldi-ai/gowa-bot/commands/owner"
	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type BotClient struct {
	Client                *whatsmeow.Client
	Registry              *lib.CommandRegistry
	Logger                *helper.Logger
	Cache                 *helper.Cache
	EphemeralHelper       *helper.EphemeralHelper
	JadibotSessionManager *helper.JadibotSessionManager
	DBManager             *helper.DatabaseManager
	Dispatcher            *lib.Dispatcher
	Activity              *helper.ActivityLog
	Owners                map[string]bool
	SelfMode              bool
	IsMainBot             bool
	Prefixes              []string
	StartedAt             time.Time
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
		Activity:              helper.NewActivityLog(50),
		Owners:                owners,
		SelfMode:              config.SelfMode,
		IsMainBot:             config.IsMainBot,
		JadibotSessionManager: config.JadibotSessionManager,
		DBManager:             config.DBManager,
		Prefixes:              prefixes,
		StartedAt:             time.Now(),
	}

	botClient.EphemeralHelper = helper.NewEphemeralHelper(nil, 5*time.Minute)

	return botClient
}

func (b *BotClient) SetClient(client *whatsmeow.Client) {
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

	if !evt.Info.Timestamp.IsZero() && evt.Info.Timestamp.Before(b.StartedAt.Add(-5*time.Second)) {
		b.Logger.Debug("Skip pending message from %s (ts=%s, started=%s)",
			evt.Info.Sender.User,
			evt.Info.Timestamp.Format(time.RFC3339),
			b.StartedAt.Format(time.RFC3339))
		return
	}

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
	case evt.Message.TemplateButtonReplyMessage != nil && evt.Message.TemplateButtonReplyMessage.SelectedID != nil:
		msg = *evt.Message.TemplateButtonReplyMessage.SelectedID
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

				b.Activity.Push(helper.ActivityEntry{
					Command:  trigger,
					Sender:   evt.Info.Sender.User,
					PushName: evt.Info.PushName,
					ChatType: chatType,
					IsOwner:  isOwner,
				})

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

	b.Activity.Push(helper.ActivityEntry{
		Command:  cmd,
		Sender:   evt.Info.Sender.User,
		PushName: evt.Info.PushName,
		ChatType: chatType,
		IsOwner:  isOwner,
	})

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

	case *events.GroupInfo:
		go b.handleGroupInfo(v)
	}
}

func (b *BotClient) handleGroupInfo(evt *events.GroupInfo) {
	if b.DBManager == nil || evt == nil {
		return
	}
	if len(evt.Join) == 0 && len(evt.Leave) == 0 {
		return
	}

	groupJID := evt.JID.String()
	cfg, err := b.DBManager.GetWelcomeConfig(groupJID)
	if err != nil {
		b.Logger.Warning("Failed to get welcome config: %v", err)
		return
	}

	b.mu.RLock()
	client := b.Client
	b.mu.RUnlock()
	if client == nil {
		return
	}

	var groupName, groupDesc string
	var memberCount int
	var addrMode types.AddressingMode
	if info, err := client.GetGroupInfo(context.Background(), evt.JID); err == nil && info != nil {
		groupName = info.Name
		groupDesc = info.Topic
		memberCount = len(info.Participants)
		addrMode = info.AddressingMode
	}

	if len(evt.Join) > 0 && cfg.WelcomeEnabled && cfg.WelcomeText != "" {
		b.sendWelcomeMessage(evt.JID, evt.Join, cfg.WelcomeText, groupName, groupDesc, memberCount, addrMode, true)
	}
	if len(evt.Leave) > 0 && cfg.GoodbyeEnabled && cfg.GoodbyeText != "" {
		b.sendWelcomeMessage(evt.JID, evt.Leave, cfg.GoodbyeText, groupName, groupDesc, memberCount, addrMode, false)
	}
}

func (b *BotClient) sendWelcomeMessage(groupJID types.JID, users []types.JID, template, groupName, groupDesc string, memberCount int, addrMode types.AddressingMode, isWelcome bool) {
	b.mu.RLock()
	client := b.Client
	b.mu.RUnlock()

	for _, user := range users {
		phoneJID, lidJID := b.resolveUserJIDs(client, user)

		var mentionNumber string
		var primaryJID types.JID
		if addrMode == types.AddressingModeLID && !lidJID.IsEmpty() {
			mentionNumber = lidJID.User
			primaryJID = lidJID
		} else if !phoneJID.IsEmpty() {
			mentionNumber = phoneJID.User
			primaryJID = phoneJID
		} else if !lidJID.IsEmpty() {
			mentionNumber = lidJID.User
			primaryJID = lidJID
		} else {
			mentionNumber = user.User
			primaryJID = user
		}
		mentionTag := "@" + mentionNumber

		// Try to resolve a display name from contact store. If found, prefix
		// the mention with the name so receivers see "Name (@123)" instead of
		// a bare number when their local pushname cache is empty (typical for
		// brand-new joiners).
		displayName := b.lookupDisplayName(client, phoneJID, lidJID, user)
		if displayName != "" {
			mentionTag = displayName + " " + mentionTag
		}

		mentioned := []string{primaryJID.String()}

		text := template
		text = strings.ReplaceAll(text, "@user", mentionTag)
		text = strings.ReplaceAll(text, "@group", groupName)
		text = strings.ReplaceAll(text, "@count", fmt.Sprintf("%d", memberCount))
		text = strings.ReplaceAll(text, "@desc", groupDesc)

		b.Logger.Info("Send welcome/goodbye: group=%s user=%s lid=%s phone=%s mode=%s tag=%s",
			groupJID.String(), user.String(), lidJID.String(), phoneJID.String(), addrMode, mentionTag)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		msg := b.buildWelcomeCanvasMessage(ctx, client, groupJID, primaryJID, phoneJID, lidJID, user, displayName, groupName, memberCount, text, mentioned, isWelcome)
		_, err := b.SendMessage(ctx, groupJID, msg)
		cancel()
		if err != nil {
			b.Logger.Warning("Failed to send welcome/goodbye to %s: %v", groupJID.String(), err)
		}
	}
}

func (b *BotClient) buildWelcomeCanvasMessage(
	ctx context.Context,
	client *whatsmeow.Client,
	groupJID, primaryJID, phoneJID, lidJID, originalJID types.JID,
	displayName, groupName string,
	memberCount int,
	caption string,
	mentioned []string,
	isWelcome bool,
) *waE2E.Message {
	textOnly := func() *waE2E.Message {
		return &waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text: proto.String(caption),
				ContextInfo: &waE2E.ContextInfo{
					MentionedJID: mentioned,
				},
			},
		}
	}

	if client == nil {
		return textOnly()
	}

	pfpURL := ""
	for _, jid := range []types.JID{primaryJID, phoneJID, lidJID, originalJID} {
		if jid.IsEmpty() {
			continue
		}
		info, err := client.GetProfilePictureInfo(ctx, jid, nil)
		if err == nil && info != nil && info.URL != "" {
			pfpURL = info.URL
			break
		}
	}

	groupIconURL := ""
	if !groupJID.IsEmpty() {
		if info, err := client.GetProfilePictureInfo(ctx, groupJID, nil); err == nil && info != nil {
			groupIconURL = info.URL
		}
	}

	userName := displayName
	if userName == "" {
		userName = "@" + primaryJID.User
	}

	imgBytes, err := helper.GenerateWelcomeCanvas(helper.CanvasParams{
		IsWelcome:    isWelcome,
		UserName:     userName,
		GroupName:    groupName,
		MemberCount:  memberCount,
		PfpURL:       pfpURL,
		GroupIconURL: groupIconURL,
	})
	if err != nil {
		b.Logger.Warning("Failed to generate canvas: %v", err)
		return textOnly()
	}

	uploadResp, err := client.Upload(ctx, imgBytes, whatsmeow.MediaImage)
	if err != nil {
		b.Logger.Warning("Failed to upload canvas: %v", err)
		return textOnly()
	}

	return &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			Caption:       proto.String(caption),
			Mimetype:      proto.String("image/jpeg"),
			URL:           &uploadResp.URL,
			DirectPath:    &uploadResp.DirectPath,
			MediaKey:      uploadResp.MediaKey,
			FileEncSHA256: uploadResp.FileEncSHA256,
			FileSHA256:    uploadResp.FileSHA256,
			FileLength:    &uploadResp.FileLength,
			ContextInfo: &waE2E.ContextInfo{
				MentionedJID: mentioned,
			},
		},
	}
}

func (b *BotClient) lookupDisplayName(client *whatsmeow.Client, jids ...types.JID) string {
	if client == nil || client.Store == nil || client.Store.Contacts == nil {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	for _, jid := range jids {
		if jid.IsEmpty() {
			continue
		}
		info, err := client.Store.Contacts.GetContact(ctx, jid)
		if err != nil || !info.Found {
			continue
		}
		if info.FullName != "" {
			return info.FullName
		}
		if info.PushName != "" {
			return info.PushName
		}
		if info.FirstName != "" {
			return info.FirstName
		}
	}
	return ""
}

func (b *BotClient) resolveUserJIDs(client *whatsmeow.Client, user types.JID) (phone, lid types.JID) {
	if client == nil || client.Store == nil {
		return user, types.JID{}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch user.Server {
	case types.DefaultUserServer:
		phone = user
		if client.Store.LIDs != nil {
			if resolved, err := client.Store.LIDs.GetLIDForPN(ctx, user); err == nil && !resolved.IsEmpty() {
				lid = resolved
			}
		}
	case types.HiddenUserServer, types.HostedLIDServer:
		lid = user
		if client.Store.LIDs != nil {
			if resolved, err := client.Store.LIDs.GetPNForLID(ctx, user); err == nil && !resolved.IsEmpty() {
				phone = resolved
			}
		}
	default:
		phone = user
	}
	return
}
