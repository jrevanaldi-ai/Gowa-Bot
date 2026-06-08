package general

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

const (
	menuTitle        = "GOWA-BOT"
	menuDescription  = "WhatsApp Bot with Gowa Library"
	menuLibraryURL   = "https://github.com/tulir/whatsmeow"
	menuSourceURL    = "https://github.com/jrevanaldi-ai/Gowa-Bot"
	menuDashboardURL = "dash.astralune.cv"
	menuThumbnailURL = "https://camo.githubusercontent.com/bf1451d500e2f05c58357170a54a224f8c0531e5af665e89df7df1fcbbc35ebd/68747470733a2f2f66696c65732e636174626f782e6d6f652f31786e7a33382e6a7067"
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
	sort.Slice(tags, func(i, j int) bool {
		return strings.ToLower(tags[i]) < strings.ToLower(tags[j])
	})

	var menuBuilder strings.Builder
	menuBuilder.WriteString("*GOWA-BOT*\n")
	menuBuilder.WriteString(menuDescription)
	menuBuilder.WriteString("\n\n")
	menuBuilder.WriteString("Library : ")
	menuBuilder.WriteString(menuLibraryURL)
	menuBuilder.WriteString("\nSource  : ")
	menuBuilder.WriteString(menuSourceURL)
	menuBuilder.WriteString("\n\n")

	for _, tag := range tags {
		tagCommands := commandsByTag[tag]
		if len(tagCommands) == 0 {
			continue
		}

		tagName := strings.ToUpper(tag)
		menuBuilder.WriteString(fmt.Sprintf("%s:\n", tagName))

		sort.Slice(tagCommands, func(i, j int) bool {
			return strings.ToLower(tagCommands[i].Cmd) < strings.ToLower(tagCommands[j].Cmd)
		})

		for _, cmd := range tagCommands {
			display := formatCommand(cmd)
			menuBuilder.WriteString(fmt.Sprintf("- %s\n", display))
		}

		menuBuilder.WriteString("\n")
	}

	menuBuilder.WriteString(fmt.Sprintf("> Dashboard: %s\n", menuDashboardURL))

	caption := strings.TrimRight(menuBuilder.String(), "\n")

	imgBytes, thumbBytes, thumbW, thumbH := helper.FetchImageFull(menuThumbnailURL)
	if len(imgBytes) == 0 {
		fallback := helper.CreateSimpleReply(caption, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String())
		_, err := ctx.SendMessage(fallback)
		if err != nil {
			return fmt.Errorf("failed to send menu: %w", err)
		}
		return nil
	}

	upload, err := ctx.Client.Upload(context.Background(), imgBytes, whatsmeow.MediaImage)
	if err != nil {
		return fmt.Errorf("failed to upload menu image: %w", err)
	}

	imgMsg := &waE2E.ImageMessage{
		URL:               proto.String(upload.URL),
		DirectPath:        proto.String(upload.DirectPath),
		Mimetype:          proto.String("image/jpeg"),
		Caption:           proto.String(caption),
		FileSHA256:        upload.FileSHA256,
		FileEncSHA256:     upload.FileEncSHA256,
		FileLength:        proto.Uint64(upload.FileLength),
		MediaKey:          upload.MediaKey,
		MediaKeyTimestamp: proto.Int64(time.Now().Unix()),
		ContextInfo: &waE2E.ContextInfo{
			StanzaID:    proto.String(ctx.MessageID),
			Participant: proto.String(ctx.Sender.String()),
		},
	}
	if len(thumbBytes) > 0 {
		imgMsg.JPEGThumbnail = thumbBytes
		if thumbW > 0 && thumbH > 0 {
			imgMsg.Height = proto.Uint32(thumbH)
			imgMsg.Width = proto.Uint32(thumbW)
		}
	}

	_, err = ctx.SendMessage(&waE2E.Message{ImageMessage: imgMsg})
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
