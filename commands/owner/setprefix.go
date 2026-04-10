package owner

import (
	"fmt"
	"strings"

	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

var SetprefixMetadata = &lib.CommandMetadata{
	Cmd:       "setprefix",
	Tag:       "owner",
	Desc:      "Set prefix bot (support multiple prefix)",
	Example:   ".setprefix . ! /",
	Hidden:    false,
	OwnerOnly: true,
	Alias:     []string{"setpref", "prefix"},
}

func SetprefixHandler(ctx *lib.CommandContext) error {

	if !ctx.IsOwner {
		message := "❌ Command ini hanya untuk owner!"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	if len(ctx.Args) == 0 {

		currentPrefixes := ctx.BotClient.GetPrefixes()

		prefixList := strings.Join(currentPrefixes, ", ")
		message := fmt.Sprintf(
			"*📋 Prefix Bot Saat Ini*\n\n"+
				"┌─⦿ *Active Prefixes*\n"+
				"│ • %s\n"+
				"└──────────────\n\n"+
				"*📝 Usage:*\n"+
				"• `.setprefix .` - Set satu prefix\n"+
				"• `.setprefix . ! /` - Set multiple prefix\n\n"+
				"*💡 Tips:*\n"+
				"• Gunakan spasi untuk memisahkan prefix\n"+
				"• Prefix yang didukung: . ! / # & dll",
			prefixList,
		)
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	newPrefixes := ctx.Args


	for _, prefix := range newPrefixes {
		if len(prefix) > 5 {
			message := "❌ *Prefix terlalu panjang!*\n\n" +
				"Maksimal 5 karakter per prefix."
			_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
			return err
		}
	}


	ctx.BotClient.SetPrefixes(newPrefixes)


	prefixList := strings.Join(newPrefixes, ", ")
	message := fmt.Sprintf(
		"*✅ Prefix Berhasil Diubah*\n\n"+
			"┌─⦿ *New Prefixes*\n"+
			"│ • %s\n"+
			"└──────────────\n\n"+
			"Sekarang bot akan merespon dengan salah satu prefix di atas.",
		prefixList,
	)

	_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return err
}

