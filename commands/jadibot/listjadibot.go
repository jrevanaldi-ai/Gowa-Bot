package jadibot

import (
	"fmt"
	"time"

	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

// ListJadibotMetadata adalah metadata untuk command listjadibot
var ListJadibotMetadata = &lib.CommandMetadata{
	Cmd:       "listjadibot",
	Tag:       "jadibot",
	Desc:      "Lihat daftar semua jadibot yang telah dibuat",
	Example:   ".listjadibot",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"listjb", "jadibotlist"},
}

// ListJadibotHandler menangani command listjadibot
func ListJadibotHandler(ctx *lib.CommandContext) error {
	// Cek apakah SessionManager tersedia
	if ctx.JadibotSessionManager == nil {
		message := "❌ *Fitur Jadibot belum diaktifkan!*\n\n" +
			"_SessionManager belum diinisialisasi. Hubungi admin bot._"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Get jadibot by owner
	jadibots, err := ctx.JadibotSessionManager.GetJadibotByOwner(ctx.Sender.String())
	if err != nil {
		message := "❌ *Gagal mengambil data jadibot!*\n\n" +
			fmt.Sprintf("┌─⦿ *Error*\n│ • %v\n└──────────────", err)
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Jika tidak ada jadibot
	if len(jadibots) == 0 {
		message := "*📋 Daftar Jadibot*\n\n" +
			"┌─⦿ *Status*\n" +
			"│ • Total Jadibot: 0\n" +
			"└──────────────\n\n" +
			"_Belum ada jadibot yang dibuat._\n\n" +
			"*📝 Cara membuat jadibot:*\n" +
			"• `.jadibot <nomor_telepon>` - Buat jadibot baru\n\n" +
			"*📖 Command Jadibot:*\n" +
			"• `.jadibot` - Buat bot baru\n" +
			"• `.listjadibot` - Lihat daftar jadibot\n" +
			"• `.stopjadibot <id>` - Hentikan jadibot\n" +
			"• `.pausejadibot <id>` - Pause jadibot\n" +
			"• `.resumejadibot <id>` - Resume jadibot"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Format daftar jadibot
	message := "*📋 Daftar Jadibot Anda*\n\n" +
		fmt.Sprintf("┌─⦿ *Total: %d bot*\n\n", len(jadibots))

	// Counter untuk statistik
	var activeCount, pausedCount, stoppedCount int

	for i, bot := range jadibots {
		// Update counter
		switch bot.Status {
		case "active":
			activeCount++
		case "paused":
			pausedCount++
		case "stopped":
			stoppedCount++
		}

		message += fmt.Sprintf("*%d. Jadibot ID: `%s`*\n", i+1, bot.ID) +
			fmt.Sprintf("   • Nomor: %s\n", bot.PhoneNumber) +
			fmt.Sprintf("   • Status: %s\n", formatJadibotStatus(bot.Status))

		// Tampilkan uptime jika aktif
		if bot.Status == "active" && bot.StartedAt != nil {
			if startTime, ok := bot.StartedAt.(time.Time); ok {
				uptime := time.Since(startTime)
				hours := int(uptime.Hours())
				minutes := int(uptime.Minutes()) % 60
				message += fmt.Sprintf("   • Uptime: %s\n", formatUptime(hours, minutes, 0))
			}
		}

		// Command yang tersedia untuk setiap bot
		message += "   • Command:\n"
		if bot.Status == "active" {
			message += fmt.Sprintf("     - `.stopjadibot %s`\n", bot.ID) +
				fmt.Sprintf("     - `.pausejadibot %s`\n", bot.ID)
		} else if bot.Status == "paused" {
			message += fmt.Sprintf("     - `.resumejadibot %s`\n", bot.ID) +
				fmt.Sprintf("     - `.stopjadibot %s`\n", bot.ID)
		} else {
			message += fmt.Sprintf("     - `.resumejadibot %s`\n", bot.ID)
		}

		message += "\n"
	}

	// Statistik
	message += "└──────────────\n\n" +
		"┌─⦿ *Statistik*\n" +
		fmt.Sprintf("│ • 🟢 Aktif: %d\n", activeCount) +
		fmt.Sprintf("│ • ⏸️ Paused: %d\n", pausedCount) +
		fmt.Sprintf("│ • ⏹️ Stopped: %d\n", stoppedCount) +
		"└──────────────\n\n" +
		"*💡 Tips:*\n" +
		"• Jadibot aktif akan otomatis merespon command\n" +
		"• Pause jika tidak digunakan untuk hemat resource\n" +
		"• Stop hanya jika ingin mematikan sementara"

	_, err = ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return err
}

// formatJadibotStatus memformat status jadibot dengan emoji
func formatJadibotStatus(status string) string {
	switch status {
	case "active":
		return "🟢 Aktif"
	case "paused":
		return "⏸️ Paused"
	case "stopped":
		return "⏹️ Berhenti"
	default:
		return "❓ Tidak Diketahui"
	}
}

// formatUptime memformat uptime menjadi string readable
func formatUptime(hours, minutes, seconds int) string {
	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}
