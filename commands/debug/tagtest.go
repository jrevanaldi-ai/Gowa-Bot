package debug

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
	"go.mau.fi/whatsmeow/proto/waE2E"
)

var TagTestMetadata = &lib.CommandMetadata{
	Cmd:       "tagtest",
	Tag:       "debug",
	Desc:      "Test mention/tag user di group (cek apakah jadi tag atau plain text)",
	Example:   ".tagtest",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"tt"},
}

func TagTestHandler(ctx *lib.CommandContext) error {
	if !ctx.IsGroup {
		_, err := ctx.SendMessage(helper.CreateSimpleReply(
			"Command ini hanya di group.", ctx.MessageID, ctx.Sender.String(), ctx.Chat.String(),
		))
		return err
	}

	sender := ctx.Sender
	mentionNumber := sender.User
	mentionTag := "@" + mentionNumber

	mode := "unknown"
	if info, err := ctx.Client.GetGroupInfo(ctx.Ctx, ctx.Chat); err == nil && info != nil {
		if info.AddressingMode != "" {
			mode = string(info.AddressingMode)
		} else {
			mode = "(default/pn)"
		}
	}

	var ownPN, ownLID string
	if ctx.Client != nil && ctx.Client.Store != nil {
		if ctx.Client.Store.ID != nil {
			ownPN = ctx.Client.Store.ID.String()
		}
		ownLID = ctx.Client.Store.LID.String()
	}

	text := fmt.Sprintf(
		"TAG TEST\n\nHello %s — kalo ini jadi tag (highlight + bisa di-tap), berarti mention jalan.\n\nDetail:\n• Sender JID: %s\n• Mention number: %s\n• Group addressing mode: %s\n• PushName sender: %s\n• Bot own PN: %s\n• Bot own LID: %s",
		mentionTag, sender.String(), mentionNumber, mode, ctx.PushName, ownPN, ownLID,
	)

	senderNonAD := sender.ToNonAD()
	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(text),
			ContextInfo: &waE2E.ContextInfo{
				MentionedJID: []string{senderNonAD.String()},
			},
		},
	}

	_, err := ctx.SendMessage(msg)
	return err
}
