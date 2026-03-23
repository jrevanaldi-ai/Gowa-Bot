package commands

import (
	"fmt"
	"reflect"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/jrevanaldi-ai/gowa/proto/waE2E"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

// PingMetadata adalah metadata untuk command ping
var PingMetadata = &lib.CommandMetadata{
	Cmd:       "ping",
	Tag:       "utility",
	Desc:      "Cek respon bot dan latency",
	Example:   ".ping",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"p"},
}

// PingHandler menangani command ping
func PingHandler(ctx *lib.CommandContext) error {
	// Format pesan response awal
	message := fmt.Sprintf("*🏓 Pong!*\n\n"+
		"┌─⦿ *Info Bot*\n"+
		"│ • *Latency:* calculating...\n"+
		"│ • *Status:* Online ✓\n"+
		"│ • *Uptime:* %s\n"+
		"└──────────────",
		getUptime())

	// Ukur latency real-time
	start := time.Now()
	sentResp, err := ctx.SendMessage(createSimpleReply(message, ctx.MessageID, ctx.Sender.String()))
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return fmt.Errorf("failed to send ping response: %w", err)
	}

	// Extract message ID dari response menggunakan reflection
	var sentMsgID string
	respValue := reflect.ValueOf(sentResp)
	if respValue.Kind() == reflect.Struct {
		idField := respValue.FieldByName("ID")
		if idField.IsValid() {
			sentMsgID = idField.String()
		}
	}

	// Jika tidak bisa extract ID, skip edit
	if sentMsgID == "" {
		return nil
	}

	// Update pesan dengan latency real
	updatedMessage := fmt.Sprintf("*🏓 Pong!*\n\n"+
		"┌─⦿ *Info Bot*\n"+
		"│ • *Latency:* %d ms\n"+
		"│ • *Status:* Online ✓\n"+
		"│ • *Uptime:* %s\n"+
		"└──────────────",
		latency,
		getUptime())

	// Edit pesan dengan latency real
	editMsg := ctx.Client.BuildEdit(ctx.Chat, sentMsgID, &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(updatedMessage),
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:    proto.String(ctx.MessageID),
				Participant: proto.String(ctx.Sender.String()),
			},
		},
	})
	_, _ = ctx.Client.SendMessage(ctx.Ctx, ctx.Chat, editMsg)

	return nil
}

// uptime adalah waktu bot mulai berjalan
var startTime = time.Now()

// getUptime mendapatkan uptime bot
func getUptime() string {
	uptime := time.Since(startTime)

	hours := int(uptime.Hours())
	minutes := int(uptime.Minutes()) % 60
	seconds := int(uptime.Seconds()) % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}
