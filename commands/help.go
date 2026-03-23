package commands

import (
	"fmt"
	"strings"

	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

// HelpMetadata adalah metadata untuk command help
var HelpMetadata = &lib.CommandMetadata{
	Cmd:       "help",
	Tag:       "main",
	Desc:      "Tampilkan informasi detail tentang command",
	Example:   ".help ping",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"info", "command"},
}

// HelpHandler menangani command help
func HelpHandler(ctx *lib.CommandContext) error {
	// Dapatkan registry dari context
	registry, ok := ctx.Ctx.Value("registry").(*lib.CommandRegistry)
	if !ok || registry == nil {
		return fmt.Errorf("registry not found in context")
	}

	// Cek apakah ada argument
	if len(ctx.Args) == 0 {
		// Jika tidak ada argument, tampilkan menu
		return MenuHandler(ctx)
	}

	// Dapatkan command yang diminta
	cmdName := ctx.Args[0]
	meta, found := registry.GetCommand(cmdName)
	if !found {
		message := fmt.Sprintf("❌ Command *%s* tidak ditemukan!\n\n"+
			"Gunakan .menu untuk melihat daftar command yang tersedia.", cmdName)
		_, err := ctx.SendMessage(createSimpleReply(message, ctx.MessageID, ctx.Sender.String()))
		return err
	}

	// Skip menu command dari help
	if meta.Cmd == "menu" {
		return MenuHandler(ctx)
	}

	// Build help message
	var helpBuilder strings.Builder
	helpBuilder.WriteString(fmt.Sprintf("*╭──⦿ HELP: %s ⦿*\n", strings.ToUpper(meta.Cmd)))
	helpBuilder.WriteString(fmt.Sprintf("│\n"))

	// Tag/Category
	helpBuilder.WriteString(fmt.Sprintf("│  *Category:* %s\n", formatTag(meta.Tag)))

	// Description
	helpBuilder.WriteString(fmt.Sprintf("│  *Description:* %s\n", meta.Desc))

	// Command usage
	helpBuilder.WriteString(fmt.Sprintf("│  *Command:* .%s\n", meta.Cmd))

	// Aliases
	if len(meta.Alias) > 0 {
		aliases := strings.Join(meta.Alias, ", ")
		helpBuilder.WriteString(fmt.Sprintf("│  *Aliases:* .%s\n", aliases))
	}

	// Example
	if meta.Example != "" {
		helpBuilder.WriteString(fmt.Sprintf("│  *Example:* %s\n", meta.Example))
	}

	// Owner only info
	if meta.OwnerOnly {
		helpBuilder.WriteString(fmt.Sprintf("│  *Access:* Owner Only 🔒\n"))
	} else {
		helpBuilder.WriteString(fmt.Sprintf("│  *Access:* Public\n"))
	}

	helpBuilder.WriteString(fmt.Sprintf("│\n"))
	helpBuilder.WriteString(fmt.Sprintf("╰──────────────────────\n"))

	message := helpBuilder.String()

	// Kirim pesan dengan reply
	_, err := ctx.SendMessage(createSimpleReply(message, ctx.MessageID, ctx.Sender.String()))
	if err != nil {
		return fmt.Errorf("failed to send help: %w", err)
	}

	return nil
}

// formatTag memformat tag menjadi lebih readable
func formatTag(tag string) string {
	// Capitalize first letter
	if len(tag) == 0 {
		return ""
	}
	return strings.ToUpper(tag[:1]) + tag[1:]
}
