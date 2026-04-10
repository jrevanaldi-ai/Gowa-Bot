package owner

import (
	"fmt"

	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

var BangroupMetadata = &lib.CommandMetadata{
	Cmd:       "bangroup",
	Tag:       "owner",
	Desc:      "Blokir grup agar bot tidak merespon command",
	Example:   ".bangroup alasan spam",
	Hidden:    false,
	OwnerOnly: true,
	Alias:     []string{"banchat"},
}

func BangroupHandler(ctx *lib.CommandContext) error {

	if !ctx.IsOwner {
		message := "❌ Command ini hanya untuk owner!"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	if !ctx.IsGroup {
		message := "❌ Command ini hanya bisa digunakan di grup!"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	dbManager, ok := ctx.BotClient.GetDBManager().(*helper.DatabaseManager)
	if !ok || dbManager == nil {
		return fmt.Errorf("database manager tidak tersedia")
	}


	groupJID := ctx.Chat.String()


	isBanned, err := dbManager.IsBanned(groupJID, "group")
	if err != nil {
		return fmt.Errorf("failed to check ban status: %w", err)
	}

	if isBanned {
		message := "❌ Grup ini sudah di-banned!"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	reason := "Tidak ada alasan"
	if len(ctx.Args) > 0 {
		reason = joinArgs(ctx.Args)
	}


	err = dbManager.BanJID(groupJID, "group", reason, ctx.Sender.String())
	if err != nil {
		return fmt.Errorf("failed to ban group: %w", err)
	}


	message := fmt.Sprintf(
		"*✅ Grup Di-Banned*\n\n"+
			"📛 *Grup:* %s\n"+
			"📝 *Alasan:* %s\n\n"+
			"Bot tidak akan merespon command dari grup ini.\n"+
			"Gunakan `.unbangroup` untuk membuka ban.",
		ctx.Chat.String(),
		reason,
	)

	_, err = ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return err
}


var UnbangroupMetadata = &lib.CommandMetadata{
	Cmd:       "unbangroup",
	Tag:       "owner",
	Desc:      "Buka blokir grup",
	Example:   ".unbangroup",
	Hidden:    false,
	OwnerOnly: true,
	Alias:     []string{"unbanchat"},
}

func UnbangroupHandler(ctx *lib.CommandContext) error {

	if !ctx.IsOwner {
		message := "❌ Command ini hanya untuk owner!"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	if !ctx.IsGroup {
		message := "❌ Command ini hanya bisa digunakan di grup!"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	dbManager, ok := ctx.BotClient.GetDBManager().(*helper.DatabaseManager)
	if !ok || dbManager == nil {
		return fmt.Errorf("database manager tidak tersedia")
	}


	groupJID := ctx.Chat.String()


	isBanned, err := dbManager.IsBanned(groupJID, "group")
	if err != nil {
		return fmt.Errorf("failed to check ban status: %w", err)
	}

	if !isBanned {
		message := "❌ Grup ini tidak sedang di-banned!"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	err = dbManager.UnbanJID(groupJID, "group")
	if err != nil {
		return fmt.Errorf("failed to unban group: %w", err)
	}


	message := fmt.Sprintf(
		"*✅ Grup Di-Unbanned*\n\n"+
			"📛 *Grup:* %s\n\n"+
			"Bot sekarang akan merespon command dari grup ini.",
		ctx.Chat.String(),
	)

	_, err = ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return err
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
