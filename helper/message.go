package helper

import (
	"fmt"
	"time"

	"github.com/jrevanaldi-ai/gowa/proto/waE2E"
	"google.golang.org/protobuf/proto"
)


type ReplyConfig struct {
	ReplyToMsgID string
	SenderJID    string
	ChatJID      string
}


func CreateSimpleReply(text string, replyToMsgID string, senderJID string, chatJID string) *waE2E.Message {

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

				Expiration:              nil,
				EphemeralSettingTimestamp: nil,
				ExternalAdReply:         nil,
				ForwardingScore:         nil,
				IsForwarded:             nil,
			},
		},
	}
}


func CreateSimpleReplyLegacy(text string, replyToMsgID string, senderJID string) *waE2E.Message {
	return CreateSimpleReply(text, replyToMsgID, senderJID, senderJID)
}


func CreateReplyFromContext(text string, ctx interface{
	GetMessageID() string
	GetSenderJID() string
	GetChatJID() string
}) *waE2E.Message {
	return CreateSimpleReply(text, ctx.GetMessageID(), ctx.GetSenderJID(), ctx.GetChatJID())
}


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



func BuildReplyMessage(message *waE2E.Message, replyToMsgID string, senderJID string, chatJID string) *waE2E.Message {
	remoteJID := chatJID
	if remoteJID == "" {
		remoteJID = senderJID
	}


	switch {
	case message.Conversation != nil:

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

		if message.ExtendedTextMessage.ContextInfo == nil {
			message.ExtendedTextMessage.ContextInfo = &waE2E.ContextInfo{}
		}
		message.ExtendedTextMessage.ContextInfo.StanzaID = &replyToMsgID
		message.ExtendedTextMessage.ContextInfo.Participant = &senderJID
		message.ExtendedTextMessage.ContextInfo.RemoteJID = &remoteJID

	case message.ImageMessage != nil:

		if message.ImageMessage.ContextInfo == nil {
			message.ImageMessage.ContextInfo = &waE2E.ContextInfo{}
		}
		message.ImageMessage.ContextInfo.StanzaID = &replyToMsgID
		message.ImageMessage.ContextInfo.Participant = &senderJID
		message.ImageMessage.ContextInfo.RemoteJID = &remoteJID

	case message.VideoMessage != nil:

		if message.VideoMessage.ContextInfo == nil {
			message.VideoMessage.ContextInfo = &waE2E.ContextInfo{}
		}
		message.VideoMessage.ContextInfo.StanzaID = &replyToMsgID
		message.VideoMessage.ContextInfo.Participant = &senderJID
		message.VideoMessage.ContextInfo.RemoteJID = &remoteJID

	case message.AudioMessage != nil:




	case message.DocumentMessage != nil:

		if message.DocumentMessage.ContextInfo == nil {
			message.DocumentMessage.ContextInfo = &waE2E.ContextInfo{}
		}
		message.DocumentMessage.ContextInfo.StanzaID = &replyToMsgID
		message.DocumentMessage.ContextInfo.Participant = &senderJID
		message.DocumentMessage.ContextInfo.RemoteJID = &remoteJID

	case message.StickerMessage != nil:

		if message.StickerMessage.ContextInfo == nil {
			message.StickerMessage.ContextInfo = &waE2E.ContextInfo{}
		}
		message.StickerMessage.ContextInfo.StanzaID = &replyToMsgID
		message.StickerMessage.ContextInfo.Participant = &senderJID
		message.StickerMessage.ContextInfo.RemoteJID = &remoteJID
	}

	return message
}


func CreateImageReply(url string, directPath string, sha256 []byte, encSha256 []byte, fileLength uint64, mediaKey []byte, replyToMsgID string, senderJID string) *waE2E.Message {
	return &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			URL:           proto.String(url),
			DirectPath:    proto.String(directPath),
			Mimetype:      proto.String("image/png"),
			FileSHA256:    sha256,
			FileEncSHA256: encSha256,
			FileLength:    proto.Uint64(fileLength),
			MediaKey:      mediaKey,
			MediaKeyTimestamp: proto.Int64(time.Now().Unix()),
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:    &replyToMsgID,
				Participant: &senderJID,
			},
		},
	}
}

