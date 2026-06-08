package maker

import (
	"fmt"
	"strings"

	"go.mau.fi/whatsmeow/proto/waE2E"

	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

var StickerMetadata = &lib.CommandMetadata{
	Cmd:       "sticker",
	Tag:       "maker",
	Desc:      "Ubah gambar jadi sticker (kirim/reply gambar dengan caption .s)",
	Example:   ".s (kirim gambar dengan caption, atau reply gambar)",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"s", "stiker"},
}

func StickerHandler(ctx *lib.CommandContext) error {
	imgMsg := extractImageMessage(ctx.GetRawMessage())
	if imgMsg == nil {
		msg := "Kirim gambar dengan caption *.s* atau reply gambar dengan *.s* untuk membuat sticker."
		_, err := ctx.SendMessage(helper.CreateSimpleReply(msg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	imgData, err := ctx.Client.Download(ctx.Ctx, imgMsg)
	if err != nil {
		errMsg := fmt.Sprintf("Gagal download gambar.\n\nError: %s", err.Error())
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	author := strings.TrimSpace(ctx.PushName)
	if author == "" {
		author = ctx.Sender.User
	}

	stickerMsg, err := helper.BuildStickerFromImage(
		ctx.Ctx,
		ctx.Client,
		imgData,
		"Gowa-Bot",
		"@"+author,
		ctx.MessageID,
		ctx.Sender.String(),
		ctx.Chat.String(),
	)
	if err != nil {
		errMsg := fmt.Sprintf("Gagal buat sticker.\n\nError: %s", err.Error())
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	if _, err := ctx.SendMessage(stickerMsg); err != nil {
		return fmt.Errorf("failed to send sticker: %w", err)
	}
	return nil
}

func extractImageMessage(m *waE2E.Message) *waE2E.ImageMessage {
	if m == nil {
		return nil
	}
	if img := m.GetImageMessage(); img != nil {
		return img
	}
	if ext := m.GetExtendedTextMessage(); ext != nil {
		if ci := ext.GetContextInfo(); ci != nil {
			if qm := ci.GetQuotedMessage(); qm != nil {
				if img := qm.GetImageMessage(); img != nil {
					return img
				}
			}
		}
	}
	return nil
}
