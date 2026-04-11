package utility

import (
	"fmt"

	"github.com/jrevanaldi-ai/gowa"
	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

var ThumbnailMetadata = &lib.CommandMetadata{
	Cmd:       "thumbnail",
	Tag:       "utility",
	Desc:      "Generate thumbnail keren untuk bot",
	Example:   ".thumbnail menu atau .thumb welcome",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"thumb", "card"},
}

func ThumbnailHandler(ctx *lib.CommandContext) error {

	if len(ctx.Args) == 0 {
		message := "*🎨 Thumbnail Generator*\n\n" +
			"┌─⦿ *Available Types*\n" +
			"│ • `menu` - Menu thumbnail\n" +
			"│ • `welcome` - Welcome card\n" +
			"│ • `info <title> <message>` - Info card\n" +
			"└──────────────\n\n" +
			"*📝 Contoh:*\n" +
			"• `.thumbnail menu`\n" +
			"• `.thumbnail welcome`\n" +
			"• `.thumbnail info Test Bot Hello`"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	thumbType := ctx.Args[0]

	var imageData []byte
	var err error

	switch thumbType {
	case "menu":
		imageData, err = helper.CreateMenuCard()
	case "welcome":
		imageData, err = helper.CreateWelcomeCard(ctx.PushName, ctx.Chat.String())
	case "info":
		if len(ctx.Args) < 3 {
			message := "❌ *Usage:*\n`.thumbnail info <title> <message>`\n\nContoh: `.thumbnail info Test Bot Hello`"
			_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
			return err
		}
		title := ctx.Args[1]
		msg := joinArgs(ctx.Args[2:])
		imageData, err = helper.CreateInfoCard(title, msg)
	default:
		message := fmt.Sprintf("❌ *Type tidak dikenal!*\n\nType tersedia: menu, welcome, info")
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	if err != nil {
		errorMsg := "❌ *Gagal generate thumbnail!*\n\n" +
			"┌─⦿ *Error*\n" +
			fmt.Sprintf("│ • %s\n", err.Error()) +
			"└──────────────"
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}


	uploadResp, err := ctx.Client.Upload(ctx.Ctx, imageData, gowa.MediaImage)
	if err != nil {
		return fmt.Errorf("failed to upload image: %w", err)
	}


	imageMsg := helper.CreateImageReply(uploadResp.URL, uploadResp.DirectPath, uploadResp.FileSHA256, uploadResp.FileEncSHA256, uploadResp.FileLength, uploadResp.MediaKey, ctx.MessageID, ctx.Sender.String())

	_, err = ctx.SendMessage(imageMsg)
	if err != nil {
		return fmt.Errorf("failed to send image: %w", err)
	}

	return nil
}


func joinArgs(args []string) string {
	result := ""
	for i, arg := range args {
		if i > 0 {
			result += " "
		}
		result += arg
	}
	return result
}
