package general

import (
	"fmt"
	"strings"

	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)


var HelpMetadata = &lib.CommandMetadata{
	Cmd:       "help",
	Tag:       "main",
	Desc:      "Tampilkan informasi detail tentang command",
	Example:   ".help ping",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"info", "command"},
}


func HelpHandler(ctx *lib.CommandContext) error {

	registry, ok := ctx.Ctx.Value("registry").(*lib.CommandRegistry)
	if !ok || registry == nil {
		return fmt.Errorf("registry not found in context")
	}


	if len(ctx.Args) == 0 {

		return MenuHandler(ctx)
	}


	cmdName := ctx.Args[0]
	meta, found := registry.GetCommand(cmdName)
	if !found {
		message := fmt.Sprintf("❌ Command *%s* tidak ditemukan!\n\n"+
			"Gunakan .menu untuk melihat daftar command yang tersedia.", cmdName)
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	if meta.Cmd == "menu" {
		return MenuHandler(ctx)
	}


	var helpBuilder strings.Builder
	helpBuilder.WriteString(fmt.Sprintf("*╭──⦿ HELP: %s ⦿*\n", strings.ToUpper(meta.Cmd)))
	helpBuilder.WriteString(fmt.Sprintf("│\n"))


	helpBuilder.WriteString(fmt.Sprintf("│  *Category:* %s\n", formatTag(meta.Tag)))


	helpBuilder.WriteString(fmt.Sprintf("│  *Description:* %s\n", meta.Desc))


	helpBuilder.WriteString(fmt.Sprintf("│  *Command:* .%s\n", meta.Cmd))


	if len(meta.Alias) > 0 {
		aliases := strings.Join(meta.Alias, ", ")
		helpBuilder.WriteString(fmt.Sprintf("│  *Aliases:* .%s\n", aliases))
	}


	if meta.Example != "" {
		helpBuilder.WriteString(fmt.Sprintf("│  *Example:* %s\n", meta.Example))
	}


	if meta.OwnerOnly {
		helpBuilder.WriteString(fmt.Sprintf("│  *Access:* Owner Only 🔒\n"))
	} else {
		helpBuilder.WriteString(fmt.Sprintf("│  *Access:* Public\n"))
	}

	helpBuilder.WriteString(fmt.Sprintf("│\n"))
	helpBuilder.WriteString(fmt.Sprintf("╰──────────────────────\n"))

	message := helpBuilder.String()


	_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	if err != nil {
		return fmt.Errorf("failed to send help: %w", err)
	}

	return nil
}


func formatTag(tag string) string {

	if len(tag) == 0 {
		return ""
	}
	return strings.ToUpper(tag[:1]) + tag[1:]
}
