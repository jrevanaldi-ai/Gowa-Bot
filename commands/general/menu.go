package general

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jrevanaldi-ai/gowa/proto/waE2E"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)


func createReplyMessage(text string, replyToMsgID string, senderJID string, externalAdReply *waE2E.ContextInfo_ExternalAdReplyInfo) *waE2E.Message {

	remoteJID := senderJID

	return &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: &text,
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:                &replyToMsgID,
				Participant:             &senderJID,
				RemoteJID:               &remoteJID,
				ExternalAdReply:         externalAdReply,

				Expiration:              nil,
				EphemeralSettingTimestamp: nil,
				ForwardingScore:         nil,
				IsForwarded:             nil,
			},
		},
	}
}


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
	menuBuilder.WriteString("*╭───⦿ GOWA-BOT ⦿───*\n")
	menuBuilder.WriteString("│\n")


	for _, tag := range tags {
		tagCommands := commandsByTag[tag]
		if len(tagCommands) == 0 {
			continue
		}


		tagName := strings.ToUpper(tag)
		menuBuilder.WriteString(fmt.Sprintf("│ *%s:*\n", tagName))


		sort.Slice(tagCommands, func(i, j int) bool {
			return tagCommands[i].Cmd < tagCommands[j].Cmd
		})

		for _, cmd := range tagCommands {
			display := formatCommand(cmd)
			menuBuilder.WriteString(fmt.Sprintf("│   • %s\n", display))
		}

		menuBuilder.WriteString("│\n")
	}


	menuBuilder.WriteString("╰────────────────\n")

	message := menuBuilder.String()


	externalAdReply := createExternalAdReply(
		"GOWA-BOT",
		"WhatsApp Bot with Gowa Library",
		"https://raw.githubusercontent.com/jrevanaldi-ai/Images/main/Gemini_Generated_Image_pmo129pmo129pmo1.png",
		"https://github.com/jrevanaldi-ai/gowa",
	)


	replyMsg := createReplyMessage(message, ctx.MessageID, ctx.Sender.String(), externalAdReply)


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
