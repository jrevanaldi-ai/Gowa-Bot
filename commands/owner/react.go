package owner

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/jrevanaldi-ai/gowa/proto/waCommon"
	"github.com/jrevanaldi-ai/gowa/proto/waE2E"
	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

var ReactMetadata = &lib.CommandMetadata{
	Cmd:       "react",
	Tag:       "owner",
	Desc:      "React ke pesan tertentu dengan emoji (owner only)",
	Example:   "react ❤️ (reply pesan)",
	Hidden:    false,
	OwnerOnly: true,
	Alias:     []string{"r"},
}

func ReactHandler(ctx *lib.CommandContext) error {
	if !ctx.IsOwner {
		return nil
	}

	if len(ctx.Args) == 0 {
		message := "❌ *Masukkan emoji untuk react!*\n\n" +
			"┌─⦿ *Usage*\n" +
			"│ • `react <emoji>` - React ke pesan (harus reply)\n" +
			"│ • `react 🗑️` - Hapus react\n" +
			"└──────────────\n\n" +
			"*📝 Contoh:*\n" +
			"• `react ❤️` (reply pesan)\n" +
			"• `react 👍` (reply pesan)\n" +
			"• `react 🗑️` (hapus react)"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Cek apakah user reply pesan
	if ctx.MessageID == "" {
		message := "❌ *Harus reply pesan yang ingin di-react!*\n\n" +
			"┌─⦿ *Cara*\n" +
			"│ 1. Reply pesan yang ingin di-react\n" +
			"│ 2. Kirim `react <emoji>`\n" +
			"└──────────────"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	emoji := strings.Join(ctx.Args, " ")

	// Cek apakah user reply pesan
	targetMsgID := ctx.MessageID
	if targetMsgID == "" {
		message := "❌ *Harus reply pesan yang ingin di-react!*\n\n" +
			"┌─⦿ *Cara*\n" +
			"│ 1. Reply pesan yang ingin di-react\n" +
			"│ 2. Kirim `react <emoji>`\n" +
			"└──────────────"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Buat dan kirim reaction message
	reactionMsg := buildReactionMessage(targetMsgID, ctx.Sender.String(), emoji)
	
	fmt.Printf("[DEBUG] Sending reaction to message: %s, participant: %s, emoji: %s\n", targetMsgID, ctx.Sender.String(), emoji)
	
	_, err := ctx.Client.SendMessage(ctx.Ctx, ctx.Chat, reactionMsg)
	if err != nil {
		fmt.Printf("[DEBUG] React error: %v\n", err)
		errorMsg := fmt.Sprintf("❌ *Gagal mengirim react!*\n\n┌─⦿ *Error*\n│ • %v\n└──────────────", err)
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}
	
	fmt.Printf("[DEBUG] React sent successfully\n")

	// Konfigurasi response
	reactText := emoji
	if emoji == "🗑️" || emoji == "" {
		reactText = "React dihapus"
	}

	message := fmt.Sprintf("✅ *React berhasil dikirim!*\n\n"+
		"┌─⦿ *Info*\n"+
		"│ • React: %s\n"+
		"│ • Target: `%s`\n"+
		"└──────────────", reactText, targetMsgID)

	_, err = ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return err
}

// buildReactionMessage membangun pesan reaction
func buildReactionMessage(targetMessageID string, participantJID string, reaction string) *waE2E.Message {
	// Buat reaction message dengan target pesan
	reactionMsg := &waE2E.ReactionMessage{
		Key: &waCommon.MessageKey{
			FromMe:    proto.Bool(false), // Target pesan orang lain
			ID:        proto.String(targetMessageID),
		},
		Text:        proto.String(reaction),
		GroupingKey: proto.String(reaction),
	}
	
	// Set participant jika di group
	if participantJID != "" {
		reactionMsg.Key.Participant = proto.String(participantJID)
	}
	
	return &waE2E.Message{
		ReactionMessage: reactionMsg,
	}
}
