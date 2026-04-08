package helper

import (
	"fmt"

	"github.com/jrevanaldi-ai/gowa/proto/waE2E"
)

// ReplyConfig konfigurasi untuk reply message
type ReplyConfig struct {
	ReplyToMsgID string
	SenderJID    string
	ChatJID      string // JID dari chat/group (bukan sender)
}

// CreateSimpleReply membuat Message dengan reply sederhana (tanpa ExternalAdReply)
func CreateSimpleReply(text string, replyToMsgID string, senderJID string, chatJID string) *waE2E.Message {
	// RemoteJID untuk reply harusnya Chat JID, bukan sender JID
	remoteJID := chatJID
	if remoteJID == "" {
		remoteJID = senderJID
	}

	return &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: &text,
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:    &replyToMsgID,
				Participant: &senderJID,
				RemoteJID:   &remoteJID,
				// Clear other ContextInfo fields
				Expiration:              nil,
				EphemeralSettingTimestamp: nil,
				ExternalAdReply:         nil,
				ForwardingScore:         nil,
				IsForwarded:             nil,
			},
		},
	}
}

// CreateSimpleReplyLegacy versi lama (backward compatibility)
func CreateSimpleReplyLegacy(text string, replyToMsgID string, senderJID string) *waE2E.Message {
	return CreateSimpleReply(text, replyToMsgID, senderJID, senderJID)
}

// CreateReplyFromContext membuat reply message dari CommandContext (lebih mudah dipakai)
func CreateReplyFromContext(text string, ctx interface{
	GetMessageID() string
	GetSenderJID() string
	GetChatJID() string
}) *waE2E.Message {
	return CreateSimpleReply(text, ctx.GetMessageID(), ctx.GetSenderJID(), ctx.GetChatJID())
}

// FormatFileSize memformat ukuran file menjadi human readable
func FormatFileSize(size int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	if size < KB {
		return fmt.Sprintf("%d B", size)
	} else if size < MB {
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	} else if size < GB {
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	}
	return fmt.Sprintf("%.2f GB", float64(size)/GB)
}

// BuildReplyMessage menambahkan ContextInfo dengan reply ke pesan asli
// Fungsi ini bisa dipakai untuk semua tipe pesan (text, video, audio, document, dll)
func BuildReplyMessage(message *waE2E.Message, replyToMsgID string, senderJID string, chatJID string) *waE2E.Message {
	remoteJID := chatJID
	if remoteJID == "" {
		remoteJID = senderJID
	}

	// Tambahkan ContextInfo ke berbagai tipe pesan
	switch {
	case message.Conversation != nil:
		// Text message - convert ke ExtendedTextMessage dengan ContextInfo
		message = &waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text: message.Conversation,
				ContextInfo: &waE2E.ContextInfo{
					StanzaID:    &replyToMsgID,
					Participant: &senderJID,
					RemoteJID:   &remoteJID,
				},
			},
		}

	case message.ExtendedTextMessage != nil:
		// Extended text message - tambahkan/update ContextInfo
		if message.ExtendedTextMessage.ContextInfo == nil {
			message.ExtendedTextMessage.ContextInfo = &waE2E.ContextInfo{}
		}
		message.ExtendedTextMessage.ContextInfo.StanzaID = &replyToMsgID
		message.ExtendedTextMessage.ContextInfo.Participant = &senderJID
		message.ExtendedTextMessage.ContextInfo.RemoteJID = &remoteJID

	case message.ImageMessage != nil:
		// Image message - tambahkan ContextInfo
		if message.ImageMessage.ContextInfo == nil {
			message.ImageMessage.ContextInfo = &waE2E.ContextInfo{}
		}
		message.ImageMessage.ContextInfo.StanzaID = &replyToMsgID
		message.ImageMessage.ContextInfo.Participant = &senderJID
		message.ImageMessage.ContextInfo.RemoteJID = &remoteJID

	case message.VideoMessage != nil:
		// Video message - tambahkan ContextInfo
		if message.VideoMessage.ContextInfo == nil {
			message.VideoMessage.ContextInfo = &waE2E.ContextInfo{}
		}
		message.VideoMessage.ContextInfo.StanzaID = &replyToMsgID
		message.VideoMessage.ContextInfo.Participant = &senderJID
		message.VideoMessage.ContextInfo.RemoteJID = &remoteJID

	case message.AudioMessage != nil:
		// Audio message - AudioMessage tidak punya ContextInfo field
		// Audio messages tidak support reply di WhatsApp
		// Biarkan tanpa ContextInfo

	case message.DocumentMessage != nil:
		// Document message - tambahkan ContextInfo
		if message.DocumentMessage.ContextInfo == nil {
			message.DocumentMessage.ContextInfo = &waE2E.ContextInfo{}
		}
		message.DocumentMessage.ContextInfo.StanzaID = &replyToMsgID
		message.DocumentMessage.ContextInfo.Participant = &senderJID
		message.DocumentMessage.ContextInfo.RemoteJID = &remoteJID

	case message.StickerMessage != nil:
		// Sticker message - tambahkan ContextInfo
		if message.StickerMessage.ContextInfo == nil {
			message.StickerMessage.ContextInfo = &waE2E.ContextInfo{}
		}
		message.StickerMessage.ContextInfo.StanzaID = &replyToMsgID
		message.StickerMessage.ContextInfo.Participant = &senderJID
		message.StickerMessage.ContextInfo.RemoteJID = &remoteJID
	}

	return message
}

