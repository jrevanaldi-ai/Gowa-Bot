package helper

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/gif"
	_ "image/png"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/jrevanaldi-ai/gowa/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

type cachedThumb struct {
	Data       []byte
	Width      uint32
	Height     uint32
	FullData   []byte
	FullWidth  uint32
	FullHeight uint32
}

var thumbnailCache sync.Map

func FetchThumbnail(thumbURL string) []byte {
	t := fetchThumbnailFull(thumbURL)
	if t == nil {
		return nil
	}
	return t.Data
}

func FetchThumbnailMeta(thumbURL string) ([]byte, uint32, uint32) {
	t := fetchThumbnailFull(thumbURL)
	if t == nil {
		return nil, 0, 0
	}
	return t.Data, t.Width, t.Height
}

func FetchImageFull(imgURL string) ([]byte, []byte, uint32, uint32) {
	t := fetchThumbnailFull(imgURL)
	if t == nil {
		return nil, nil, 0, 0
	}
	return t.FullData, t.Data, t.FullWidth, t.FullHeight
}

func fetchThumbnailFull(thumbURL string) *cachedThumb {
	if thumbURL == "" {
		return nil
	}
	if cached, ok := thumbnailCache.Load(thumbURL); ok {
		if t, ok := cached.(*cachedThumb); ok {
			return t
		}
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", thumbURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4*1024*1024))
	if err != nil || len(raw) == 0 {
		return nil
	}

	processed, w, h := processThumbnail(raw)
	if processed == nil {
		return nil
	}

	fullW, fullH := imageDimensions(raw)

	thumb := &cachedThumb{
		Data:       processed,
		Width:      w,
		Height:     h,
		FullData:   raw,
		FullWidth:  fullW,
		FullHeight: fullH,
	}
	thumbnailCache.Store(thumbURL, thumb)
	return thumb
}

func imageDimensions(raw []byte) (uint32, uint32) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil {
		return 0, 0
	}
	return uint32(cfg.Width), uint32(cfg.Height)
}

func processThumbnail(raw []byte) ([]byte, uint32, uint32) {
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, 0, 0
	}

	const maxSide = 192
	bounds := img.Bounds()
	srcW, srcH := bounds.Dx(), bounds.Dy()

	dstW, dstH := srcW, srcH
	if srcW > maxSide || srcH > maxSide {
		if srcW >= srcH {
			dstW = maxSide
			dstH = srcH * maxSide / srcW
		} else {
			dstH = maxSide
			dstW = srcW * maxSide / srcH
		}
	}

	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	for y := 0; y < dstH; y++ {
		sy := bounds.Min.Y + y*srcH/dstH
		for x := 0; x < dstW; x++ {
			sx := bounds.Min.X + x*srcW/dstW
			dst.Set(x, y, img.At(sx, sy))
		}
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 70}); err != nil {
		return nil, 0, 0
	}
	return buf.Bytes(), uint32(dstW), uint32(dstH)
}

func CreateLinkPreviewReply(text, title, description, sourceURL string, thumbBytes []byte, replyToMsgID, senderJID, chatJID string) *waE2E.Message {
	return CreateLinkPreviewReplyWithSize(text, title, description, sourceURL, thumbBytes, 0, 0, replyToMsgID, senderJID, chatJID)
}

func CreateLinkPreviewReplyWithSize(text, title, description, sourceURL string, thumbBytes []byte, thumbW, thumbH uint32, replyToMsgID, senderJID, chatJID string) *waE2E.Message {
	remoteJID := chatJID
	if remoteJID == "" {
		remoteJID = senderJID
	}
	previewType := waE2E.ExtendedTextMessage_IMAGE

	ext := &waE2E.ExtendedTextMessage{
		Text:        proto.String(text),
		Title:       proto.String(title),
		Description: proto.String(description),
		MatchedText: proto.String(sourceURL),
		PreviewType: &previewType,
		ContextInfo: &waE2E.ContextInfo{
			StanzaID:    proto.String(replyToMsgID),
			Participant: proto.String(senderJID),
			RemoteJID:   proto.String(remoteJID),
		},
	}
	if len(thumbBytes) > 0 {
		ext.JPEGThumbnail = thumbBytes
		if thumbW > 0 && thumbH > 0 {
			ext.ThumbnailWidth = proto.Uint32(thumbW)
			ext.ThumbnailHeight = proto.Uint32(thumbH)
		}
	}

	return &waE2E.Message{ExtendedTextMessage: ext}
}


func ExtractMessageText(m *waE2E.Message) string {
	if m == nil {
		return ""
	}
	switch {
	case m.Conversation != nil:
		return *m.Conversation
	case m.ExtendedTextMessage != nil && m.ExtendedTextMessage.Text != nil:
		return *m.ExtendedTextMessage.Text
	case m.ImageMessage != nil && m.ImageMessage.Caption != nil:
		return *m.ImageMessage.Caption
	case m.VideoMessage != nil && m.VideoMessage.Caption != nil:
		return *m.VideoMessage.Caption
	case m.DocumentMessage != nil && m.DocumentMessage.Caption != nil:
		return *m.DocumentMessage.Caption
	case m.EphemeralMessage != nil && m.EphemeralMessage.Message != nil:
		return ExtractMessageText(m.EphemeralMessage.Message)
	case m.ViewOnceMessage != nil && m.ViewOnceMessage.Message != nil:
		return ExtractMessageText(m.ViewOnceMessage.Message)
	case m.ViewOnceMessageV2 != nil && m.ViewOnceMessageV2.Message != nil:
		return ExtractMessageText(m.ViewOnceMessageV2.Message)
	case m.DocumentWithCaptionMessage != nil && m.DocumentWithCaptionMessage.Message != nil:
		return ExtractMessageText(m.DocumentWithCaptionMessage.Message)
	}
	return ""
}

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

