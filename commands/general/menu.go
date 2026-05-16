package general

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

const (
	menuTitle        = "GOWA-BOT"
	menuDescription  = "WhatsApp Bot with Gowa Library"
	menuSourceURL    = "https://github.com/jrevanaldi-ai/gowa"
	menuThumbnailURL = "https://files.catbox.moe/1xnz38.jpg"
)


var MenuMetadata = &lib.CommandMetadata{
	Cmd:       "menu",
	Tag:       "main",
	Desc:      "Tampilkan daftar command yang tersedia",
	Example:   ".menu",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"m", "help", "h"},
}


func MenuHandler(ctx *lib.CommandContext) error {

	registry, ok := ctx.Ctx.Value("registry").(*lib.CommandRegistry)
	if !ok || registry == nil {
		return fmt.Errorf("registry not found in context")
	}


	commands := registry.GetAllCommands()


	commandsByTag := make(map[string][]*lib.CommandMetadata)
	for _, cmd := range commands {

		if cmd.Cmd == "menu" {
			continue
		}
		commandsByTag[cmd.Tag] = append(commandsByTag[cmd.Tag], cmd)
	}


	var tags []string
	for tag := range commandsByTag {
		tags = append(tags, tag)
	}
	sort.Strings(tags)


	var menuBuilder strings.Builder
	menuBuilder.WriteString(menuSourceURL)
	menuBuilder.WriteString("\n\nGOWA-BOT\n\n")


	for _, tag := range tags {
		tagCommands := commandsByTag[tag]
		if len(tagCommands) == 0 {
			continue
		}


		tagName := strings.ToUpper(tag)
		menuBuilder.WriteString(fmt.Sprintf("%s:\n", tagName))


		sort.Slice(tagCommands, func(i, j int) bool {
			return tagCommands[i].Cmd < tagCommands[j].Cmd
		})

		for _, cmd := range tagCommands {
			display := formatCommand(cmd)
			menuBuilder.WriteString(fmt.Sprintf("- %s\n", display))
		}

		menuBuilder.WriteString("\n")
	}

	message := menuBuilder.String()

	thumbBytes, thumbW, thumbH := helper.FetchThumbnailMeta(menuThumbnailURL)
	replyMsg := helper.CreateLinkPreviewReplyWithSize(
		message,
		menuTitle,
		menuDescription,
		menuSourceURL,
		thumbBytes,
		thumbW, thumbH,
		ctx.MessageID,
		ctx.Sender.String(),
		ctx.Chat.String(),
	)

	_, err := ctx.SendMessage(replyMsg)
	if err != nil {
		return fmt.Errorf("failed to send menu: %w", err)
	}

	return nil
}



func formatCommand(cmd *lib.CommandMetadata) string {
	var parts []string


	parts = append(parts, cmd.Cmd)


	if len(cmd.Alias) > 0 {
		parts = append(parts, fmt.Sprintf("(%s)", cmd.Alias[0]))
	}

	return strings.Join(parts, " ")
}
