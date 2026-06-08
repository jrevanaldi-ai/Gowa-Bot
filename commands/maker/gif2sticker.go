package maker

import (
	"fmt"
	"strings"

	"go.mau.fi/whatsmeow/proto/waE2E"

	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

var Gif2StickerMetadata = &lib.CommandMetadata{
	Cmd:       "gif2sticker",
	Tag:       "maker",
	Desc:      "Ubah GIF/video pendek jadi sticker animasi (kirim/reply video atau GIF dengan .gif2s)",
	Example:   ".gif2s (kirim GIF/video dengan caption, atau reply GIF/video)",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"gif2s", "gif2stiker", "videosticker", "vsticker"},
}

func Gif2StickerHandler(ctx *lib.CommandContext) error {
	vidMsg := extractVideoMessage(ctx.GetRawMessage())
	if vidMsg == nil {
		msg := "Kirim GIF/video pendek dengan caption *.gif2s* atau reply GIF/video dengan *.gif2s* untuk membuat sticker animasi.\n\n" +
			"Catatan:\n" +
			"- Durasi maksimal 6 detik (otomatis dipotong)\n" +
			"- Output 512x512, looping infinite"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(msg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	vidData, err := ctx.Client.Download(ctx.Ctx, vidMsg)
	if err != nil {
		errMsg := fmt.Sprintf("Gagal download video.\n\nError: %s", err.Error())
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	author := strings.TrimSpace(ctx.PushName)
	if author == "" {
		author = ctx.Sender.User
	}

	stickerMsg, err := helper.BuildAnimatedStickerFromVideo(
		ctx.Ctx,
		ctx.Client,
		vidData,
		"Gowa-Bot",
		"@"+author,
		ctx.MessageID,
		ctx.Sender.String(),
		ctx.Chat.String(),
	)
	if err != nil {
		errMsg := fmt.Sprintf("Gagal buat sticker animasi.\n\nError: %s", err.Error())
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	if _, err := ctx.SendMessage(stickerMsg); err != nil {
		return fmt.Errorf("failed to send animated sticker: %w", err)
	}
	return nil
}

func extractVideoMessage(m *waE2E.Message) *waE2E.VideoMessage {
	if m == nil {
		return nil
	}
	if v := m.GetVideoMessage(); v != nil {
		return v
	}
	if ext := m.GetExtendedTextMessage(); ext != nil {
		if ci := ext.GetContextInfo(); ci != nil {
			if qm := ci.GetQuotedMessage(); qm != nil {
				if v := qm.GetVideoMessage(); v != nil {
					return v
				}
			}
		}
	}
	return nil
}
