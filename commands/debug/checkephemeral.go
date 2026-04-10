package debug

import (
	"context"
	"fmt"

	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)


var CheckEphemeralMetadata = &lib.CommandMetadata{
	Cmd:       "checkephemeral",
	Tag:       "debug",
	Desc:      "Cek status ephemeral di group ini",
	Example:   ".checkephemeral",
	Hidden:    true,
	OwnerOnly: true,
	Alias:     []string{"ce"},
}


func CheckEphemeralHandler(ctx *lib.CommandContext) error {

	if !ctx.IsGroup {
		message := "❌ Command ini hanya bisa digunakan di group!"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	type ephemeralChecker interface {
		GetEphemeralConfig(ctx context.Context, chatID string) (*EphemeralConfig, error)
	}


	config, err := getEphemeralConfig(ctx)
	if err != nil {
		message := fmt.Sprintf("❌ Error: %v", err)
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	var message string
	if config.IsEphemeral && config.DisappearingTimer > 0 {
		timer := formatTimer(config.DisappearingTimer)
		message = fmt.Sprintf("*✅ Ephemeral Status*\n\n"+
			"┌─⦿ *Info Group*\n"+
			"│ • *Group:* %s\n"+
			"│ • *Ephemeral:* Enabled ✓\n"+
			"│ • *Timer:* %s\n"+
			"└──────────────",
			ctx.Chat.String(),
			timer)
	} else {
		message = fmt.Sprintf("*❌ Ephemeral Status*\n\n"+
			"┌─⦿ *Info Group*\n"+
			"│ • *Group:* %s\n"+
			"│ • *Ephemeral:* Disabled ✗\n"+
			"│ • *Timer:* Off\n"+
			"└──────────────",
			ctx.Chat.String())
	}

	_, err = ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return err
}


type EphemeralConfig struct {
	IsEphemeral       bool
	DisappearingTimer uint32
}


func getEphemeralConfig(ctx *lib.CommandContext) (*EphemeralConfig, error) {

	groupInfo, err := ctx.Client.GetGroupInfo(ctx.Ctx, ctx.Chat)
	if err != nil {
		return &EphemeralConfig{}, fmt.Errorf("failed to get group info: %w", err)
	}

	return &EphemeralConfig{
		IsEphemeral:       groupInfo.IsEphemeral,
		DisappearingTimer: groupInfo.DisappearingTimer,
	}, nil
}


func formatTimer(seconds uint32) string {
	if seconds == 0 {
		return "Off"
	}

	minutes := seconds / 60
	hours := minutes / 60
	days := hours / 24

	if days > 0 {
		return fmt.Sprintf("%d hari", days)
	} else if hours > 0 {
		return fmt.Sprintf("%d jam", hours)
	} else if minutes > 0 {
		return fmt.Sprintf("%d menit", minutes)
	}
	return fmt.Sprintf("%d detik", seconds)
}
