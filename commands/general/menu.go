package general

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jrevanaldi-ai/gowa/proto/waE2E"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

// createReplyMessage membuat Message dengan reply ke pesan asli
func createReplyMessage(text string, replyToMsgID string, senderJID string, externalAdReply *waE2E.ContextInfo_ExternalAdReplyInfo) *waE2E.Message {
	// RemoteJID untuk reply
	remoteJID := senderJID

	return &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: &text,
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:                &replyToMsgID,
				Participant:             &senderJID,
				RemoteJID:               &remoteJID,
				ExternalAdReply:         externalAdReply,
				// Clear other ContextInfo fields
				Expiration:              nil,
				EphemeralSettingTimestamp: nil,
				ForwardingScore:         nil,
				IsForwarded:             nil,
			},
		},
	}
}

// createExternalAdReply membuat ContextInfo dengan ExternalAdReply
func createExternalAdReply(title, body, thumbnailURL, sourceURL string) *waE2E.ContextInfo_ExternalAdReplyInfo {
	mediaType := waE2E.ContextInfo_ExternalAdReplyInfo_IMAGE
	adType := waE2E.ContextInfo_ExternalAdReplyInfo_CTWA
	showAdAttribution := true
	renderLargerThumbnail := true

	return &waE2E.ContextInfo_ExternalAdReplyInfo{
		Title:                 &title,
		Body:                  &body,
		MediaType:             &mediaType,
		ThumbnailURL:          &thumbnailURL,
		SourceURL:             &sourceURL,
		ShowAdAttribution:     &showAdAttribution,
		RenderLargerThumbnail: &renderLargerThumbnail,
		AdType:                &adType,
	}
}

// MenuMetadata adalah metadata untuk command menu
var MenuMetadata = &lib.CommandMetadata{
	Cmd:       "menu",
	Tag:       "main",
	Desc:      "Tampilkan daftar command yang tersedia",
	Example:   ".menu",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"m", "help", "h"},
}

// MenuHandler menangani command menu
func MenuHandler(ctx *lib.CommandContext) error {
	// Dapatkan registry dari context (akan di-set di client)
	registry, ok := ctx.Ctx.Value("registry").(*lib.CommandRegistry)
	if !ok || registry == nil {
		return fmt.Errorf("registry not found in context")
	}

	// Dapatkan semua command yang tidak hidden
	commands := registry.GetAllCommands()

	// Group commands by tag
	commandsByTag := make(map[string][]*lib.CommandMetadata)
	for _, cmd := range commands {
		// Skip menu command itself dari listing
		if cmd.Cmd == "menu" {
			continue
		}
		commandsByTag[cmd.Tag] = append(commandsByTag[cmd.Tag], cmd)
	}

	// Sort tags
	var tags []string
	for tag := range commandsByTag {
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	// Build menu message
	var menuBuilder strings.Builder
	menuBuilder.WriteString("*╭───⦿ GOWA-BOT ⦿───*\n")
	menuBuilder.WriteString("│\n")

	// Tampilkan commands per tag
	for _, tag := range tags {
		tagCommands := commandsByTag[tag]
		if len(tagCommands) == 0 {
			continue
		}

		// Format tag name
		tagName := strings.ToUpper(tag)
		menuBuilder.WriteString(fmt.Sprintf("│ *%s:*\n", tagName))

		// Sort commands by cmd name
		sort.Slice(tagCommands, func(i, j int) bool {
			return tagCommands[i].Cmd < tagCommands[j].Cmd
		})

		for _, cmd := range tagCommands {
			display := formatCommand(cmd)
			menuBuilder.WriteString(fmt.Sprintf("│   • %s\n", display))
		}

		menuBuilder.WriteString("│\n")
	}

	// Footer
	menuBuilder.WriteString("╰────────────────\n")

	message := menuBuilder.String()

	// Buat ExternalAdReply
	externalAdReply := createExternalAdReply(
		"GOWA-BOT",
		"WhatsApp Bot with Gowa Library",
		"https://raw.githubusercontent.com/jrevanaldi-ai/Images/main/Gemini_Generated_Image_pmo129pmo129pmo1.png", // Golang mascot
		"https://github.com/jrevanaldi-ai/gowa",
	)

	// Buat pesan dengan reply
	replyMsg := createReplyMessage(message, ctx.MessageID, ctx.Sender.String(), externalAdReply)

	// Kirim pesan
	_, err := ctx.SendMessage(replyMsg)
	if err != nil {
		return fmt.Errorf("failed to send menu: %w", err)
	}

	return nil
}

// formatCommand memformat command untuk ditampilkan di menu
// Format: command (alias)
func formatCommand(cmd *lib.CommandMetadata) string {
	var parts []string

	// Add main command
	parts = append(parts, cmd.Cmd)

	// Add first alias only (hardcoded)
	if len(cmd.Alias) > 0 {
		parts = append(parts, fmt.Sprintf("(%s)", cmd.Alias[0]))
	}

	return strings.Join(parts, " ")
}
