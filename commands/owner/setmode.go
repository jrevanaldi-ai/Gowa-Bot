package owner

import (
	"fmt"
	"strings"

	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

// SetmodeMetadata adalah metadata untuk command setmode
var SetmodeMetadata = &lib.CommandMetadata{
	Cmd:       "setmode",
	Tag:       "owner",
	Desc:      "Set mode bot (self/public)",
	Example:   ".setmode self atau .setmode public",
	Hidden:    false,
	OwnerOnly: true,
	Alias:     []string{"mode"},
}

// SetmodeHandler menangani command setmode
func SetmodeHandler(ctx *lib.CommandContext) error {
	// Hanya owner yang bisa menggunakan
	if !ctx.IsOwner {
		message := "❌ Command ini hanya untuk owner!"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Cek apakah ada argument
	if len(ctx.Args) == 0 {
		message := "*📋 Mode Bot*\n\n" +
			"Usage:\n" +
			"• `.setmode self` - Aktifkan self mode\n" +
			"• `.setmode public` - Aktifkan public mode"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Parse argument
	modeArg := strings.ToLower(ctx.Args[0])

	var newMode bool
	var modeName string

	switch modeArg {
	case "self":
		newMode = true
		modeName = "Self Mode"
	case "public":
		newMode = false
		modeName = "Public Mode"
	default:
		message := "❌ Mode tidak valid!\n\n" +
			"Gunakan:\n" +
			"• `self` - untuk self mode\n" +
			"• `public` - untuk public mode"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Set mode di BotClient
	if ctx.BotClient != nil {
		ctx.BotClient.SetSelfMode(newMode)
	}

	// Format pesan sukses
	message := fmt.Sprintf("*✅ %s Aktif*\n\n"+
		"Usage:\n"+
		"• `.setmode self` - Aktifkan self mode\n"+
		"• `.setmode public` - Aktifkan public mode", modeName)

	_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return err
}
