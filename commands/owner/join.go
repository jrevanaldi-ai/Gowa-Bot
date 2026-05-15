package owner

import (
	"fmt"

	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

var JoinMetadata = &lib.CommandMetadata{
	Cmd:       "join",
	Tag:       "owner",
	Desc:      "Join ke grup WhatsApp via invite link",
	Example:   ".join https://chat.whatsapp.com/ABC123XYZ",
	Hidden:    false,
	OwnerOnly: true,
	Alias:     []string{"joingrup", "joingroup"},
}

func JoinHandler(ctx *lib.CommandContext) error {
	if !ctx.IsOwner {
		message := "Command ini hanya untuk owner."
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	var rawLink string
	if len(ctx.Args) > 0 {
		rawLink = ctx.Args[0]
	} else if ctx.ReplyMessage != nil {
		rawLink = helper.ExtractMatchingURL(ctx.ReplyMessage.Message, func(u string) bool {
			return helper.ExtractWhatsAppInviteCode(u) != ""
		})
	}

	if rawLink == "" {
		message := "Join Grup\n\n" +
			"Usage:\n" +
			"- .join <link grup>\n" +
			"- Atau reply pesan berisi link grup dengan .join\n\n" +
			"Contoh:\n" +
			"- .join https://chat.whatsapp.com/ABC123XYZ\n" +
			"- .join ABC123XYZ"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	code := helper.ExtractWhatsAppInviteCode(rawLink)
	if code == "" {
		message := "Link grup tidak valid.\n\n" +
			"Pastikan formatnya seperti:\n" +
			"https://chat.whatsapp.com/ABC123XYZ"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	groupJID, err := ctx.Client.JoinGroupWithLink(ctx.Ctx, code)
	if err != nil {
		errorMsg := fmt.Sprintf("Gagal join grup.\n\nError: %s", err.Error())
		_, sendErr := ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return sendErr
	}

	successMsg := fmt.Sprintf("Berhasil join grup.\n\nJID: %s", groupJID.String())

	_, sendErr := ctx.SendMessage(helper.CreateSimpleReply(successMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return sendErr
}

