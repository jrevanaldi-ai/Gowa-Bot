package owner

import (
	"fmt"
	"strings"

	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

var BanuserMetadata = &lib.CommandMetadata{
	Cmd:       "banuser",
	Tag:       "owner",
	Desc:      "Blokir user agar tidak bisa menggunakan command bot",
	Example:   ".banuser 6281234567890 alasan spam",
	Hidden:    false,
	OwnerOnly: true,
	Alias:     []string{"ban"},
}

func BanuserHandler(ctx *lib.CommandContext) error {

	if !ctx.IsOwner {
		message := "❌ Command ini hanya untuk owner!"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	if len(ctx.Args) == 0 {
		message := "*📋 Ban User*\n\n" +
			"Usage:\n" +
			"• `.banuser 6281234567890` - Ban user berdasarkan nomor\n" +
			"• `.banuser 6281234567890 alasan spam` - Ban dengan alasan"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	dbManager, ok := ctx.BotClient.GetDBManager().(*helper.DatabaseManager)
	if !ok || dbManager == nil {
		return fmt.Errorf("database manager tidak tersedia")
	}


	targetJID := ctx.Args[0]


	if !strings.Contains(targetJID, "@") {
		targetJID = targetJID + "@s.whatsapp.net"
	}


	isBanned, err := dbManager.IsBanned(targetJID, "user")
	if err != nil {
		return fmt.Errorf("failed to check ban status: %w", err)
	}

	if isBanned {
		message := "❌ User ini sudah di-banned!"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	reason := "Tidak ada alasan"
	if len(ctx.Args) > 1 {
		reason = joinArgs(ctx.Args[1:])
	}


	err = dbManager.BanJID(targetJID, "user", reason, ctx.Sender.String())
	if err != nil {
		return fmt.Errorf("failed to ban user: %w", err)
	}


	message := fmt.Sprintf(
		"*✅ User Di-Banned*\n\n"+
			"👤 *User:* %s\n"+
			"📝 *Alasan:* %s\n\n"+
			"User ini tidak bisa menggunakan command bot.\n"+
			"Gunakan `.unbanuser` untuk membuka ban.",
		targetJID,
		reason,
	)

	_, err = ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return err
}


var UnbanuserMetadata = &lib.CommandMetadata{
	Cmd:       "unbanuser",
	Tag:       "owner",
	Desc:      "Buka blokir user",
	Example:   ".unbanuser 6281234567890",
	Hidden:    false,
	OwnerOnly: true,
	Alias:     []string{"unban"},
}

func UnbanuserHandler(ctx *lib.CommandContext) error {

	if !ctx.IsOwner {
		message := "❌ Command ini hanya untuk owner!"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	if len(ctx.Args) == 0 {
		message := "*📋 Unban User*\n\n" +
			"Usage:\n" +
			"• `.unbanuser 6281234567890` - Unban user berdasarkan nomor"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	dbManager, ok := ctx.BotClient.GetDBManager().(*helper.DatabaseManager)
	if !ok || dbManager == nil {
		return fmt.Errorf("database manager tidak tersedia")
	}


	targetJID := ctx.Args[0]


	if !strings.Contains(targetJID, "@") {
		targetJID = targetJID + "@s.whatsapp.net"
	}


	isBanned, err := dbManager.IsBanned(targetJID, "user")
	if err != nil {
		return fmt.Errorf("failed to check ban status: %w", err)
	}

	if !isBanned {
		message := "❌ User ini tidak sedang di-banned!"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	err = dbManager.UnbanJID(targetJID, "user")
	if err != nil {
		return fmt.Errorf("failed to unban user: %w", err)
	}


	message := fmt.Sprintf(
		"*✅ User Di-Unbanned*\n\n"+
			"👤 *User:* %s\n\n"+
			"User ini sekarang bisa menggunakan command bot.",
		targetJID,
	)

	_, err = ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return err
}
