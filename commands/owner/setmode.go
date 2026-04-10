package owner

import (
	"fmt"
	"strings"

	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)


var SetmodeMetadata = &lib.CommandMetadata{
	Cmd:       "setmode",
	Tag:       "owner",
	Desc:      "Set mode bot (self/public)",
	Example:   ".setmode self atau .setmode public",
	Hidden:    false,
	OwnerOnly: true,
	Alias:     []string{"mode"},
}


func SetmodeHandler(ctx *lib.CommandContext) error {

	if !ctx.IsOwner {
		message := "❌ Command ini hanya untuk owner!"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	if len(ctx.Args) == 0 {
		message := "*📋 Mode Bot*\n\n" +
			"Usage:\n" +
			"• `.setmode self` - Aktifkan self mode\n" +
			"• `.setmode public` - Aktifkan public mode"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


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


	if ctx.BotClient != nil {
		ctx.BotClient.SetSelfMode(newMode)
	}


	message := fmt.Sprintf("*✅ %s Aktif*\n\n"+
		"Usage:\n"+
		"• `.setmode self` - Aktifkan self mode\n"+
		"• `.setmode public` - Aktifkan public mode", modeName)

	_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return err
}
