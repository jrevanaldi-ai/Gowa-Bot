package general

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/jrevanaldi-ai/gowa"
	"github.com/jrevanaldi-ai/gowa/proto/waE2E"
	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)


var GetppMetadata = &lib.CommandMetadata{
	Cmd:       "getpp",
	Tag:       "main",
	Desc:      "Ambil foto profil user (reply, tag, atau nomor)",
	Example:   ".getpp 6281234567890",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"pp", "avatar"},
}


func GetppHandler(ctx *lib.CommandContext) error {

	var targetJID string


	if ctx.ReplyMessage != nil && ctx.ReplyMessage.Sender != "" {

		targetJID = ctx.ReplyMessage.Sender
	} else if len(ctx.Mentions) > 0 {

		targetJID = ctx.Mentions[0]
	} else if len(ctx.Args) > 0 {

		phone := ctx.Args[0]

		phone = strings.TrimLeft(phone, "+")
		phone = strings.ReplaceAll(phone, "-", "")
		phone = strings.ReplaceAll(phone, " ", "")


		if !strings.HasSuffix(phone, "@s.whatsapp.net") {
			phone = phone + "@s.whatsapp.net"
		}
		targetJID = phone
	} else {

		message := "❌ Format salah!\n\n" +
			"Gunakan salah satu cara berikut:\n" +
			"• Reply pesan user\n" +
			"• Tag user: @6281234567890\n" +
			"• Masukkan nomor: .getpp 6281234567890"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	loadingMsg := "⏳ Mengambil foto profil..."
	sentMsg, err := ctx.SendMessage(helper.CreateSimpleReply(loadingMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	if err != nil {
		return fmt.Errorf("failed to send loading message: %w", err)
	}
	loadingMsgID := sentMsgToString(sentMsg)


	ctxTimeout, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()


	profilePicInfo, err := ctx.Client.GetProfilePictureInfo(ctxTimeout, lib.StringToJID(targetJID), &gowa.GetProfilePictureParams{
		Preview: false,
	})
	if err != nil {
		var errorMsg string
		if err == gowa.ErrProfilePictureUnauthorized {
			errorMsg = "❌ User ini menyembunyikan foto profilnya dari Anda 🔒"
		} else if err == gowa.ErrProfilePictureNotSet {
			errorMsg = "❌ User ini tidak memiliki foto profil"
		} else {
			errorMsg = fmt.Sprintf("❌ Gagal mengambil foto profil:\n```%v```", err)
		}


		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, loadingMsgID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	if profilePicInfo == nil || profilePicInfo.URL == "" {
		message := "❌ User ini tidak memiliki foto profil"
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(message, loadingMsgID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}


	httpClient := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := httpClient.Get(profilePicInfo.URL)
	if err != nil {
		errorMsg := fmt.Sprintf("❌ Gagal mendownload foto:\n```%v```", err)
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, loadingMsgID, ctx.Sender.String(), ctx.Chat.String()))
		return fmt.Errorf("failed to download profile picture: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorMsg := fmt.Sprintf("❌ Gagal mendownload foto (HTTP %d)", resp.StatusCode)
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, loadingMsgID, ctx.Sender.String(), ctx.Chat.String()))
		return fmt.Errorf("failed to download profile picture: HTTP %d", resp.StatusCode)
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		errorMsg := fmt.Sprintf("❌ Gagal membaca foto:\n```%v```", err)
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, loadingMsgID, ctx.Sender.String(), ctx.Chat.String()))
		return fmt.Errorf("failed to read image data: %w", err)
	}


	uploadResp, err := ctx.Client.Upload(context.Background(), imageData, gowa.MediaImage)
	if err != nil {
		errorMsg := fmt.Sprintf("❌ Gagal mengupload foto:\n```%v```", err)
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, loadingMsgID, ctx.Sender.String(), ctx.Chat.String()))
		return fmt.Errorf("failed to upload image: %w", err)
	}


	caption := fmt.Sprintf(
		"✅ *Foto Profil Berhasil Diambil!*\n\n"+
			"┌─⦿ Info User\n"+
			"│ • JID: ```%s```\n"+
			"│ • Picture ID: ```%s```\n"+
			"└──────────────\n\n"+
			"_💡 Tips: Klik gambar untuk melihat full size_",
		formatJID(targetJID),
		profilePicInfo.ID,
	)


	imageMsg := &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			URL:                 proto.String(uploadResp.URL),
			DirectPath:          proto.String(uploadResp.DirectPath),
			Mimetype:            proto.String("image/jpeg"),
			Caption:             proto.String(caption),
			FileSHA256:          uploadResp.FileSHA256,
			FileEncSHA256:       uploadResp.FileEncSHA256,
			FileLength:          proto.Uint64(uploadResp.FileLength),
			MediaKey:            uploadResp.MediaKey,
			MediaKeyTimestamp:   proto.Int64(time.Now().Unix()),
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:    proto.String(ctx.MessageID),
				Participant: proto.String(ctx.Sender.String()),
			},
		},
	}

	_, err = ctx.SendMessage(imageMsg)
	if err != nil {
		return fmt.Errorf("failed to send profile picture: %w", err)
	}

	return nil
}


func sentMsgToString(sentMsg interface{}) string {

	type Messager interface {
		GetMessageID() string
	}

	if m, ok := sentMsg.(Messager); ok {
		return m.GetMessageID()
	}

	return ""
}


func formatJID(jid string) string {

	parts := strings.Split(jid, "@")
	if len(parts) > 0 {
		return parts[0]
	}
	return jid
}
