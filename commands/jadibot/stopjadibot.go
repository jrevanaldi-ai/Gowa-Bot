package jadibot

import (
	"fmt"

	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

// StopJadibotMetadata adalah metadata untuk command stopjadibot
var StopJadibotMetadata = &lib.CommandMetadata{
	Cmd:       "stopjadibot",
	Tag:       "jadibot",
	Desc:      "Hentikan jadibot yang sedang berjalan",
	Example:   ".stopjadibot <id_jadibot>",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"stopjb", "matibot"},
}

// StopJadibotHandler menangani command stopjadibot
func StopJadibotHandler(ctx *lib.CommandContext) error {
	// Cek apakah ada argument (ID jadibot)
	if len(ctx.Args) == 0 {
		message := "*⏹️ Stop Jadibot*\n\n" +
			"┌─⦿ *Usage*\n" +
			"│ • `.stopjadibot <id_jadibot>` - Hentikan jadibot\n" +
			"└──────────────\n\n" +
			"*📋 Cara Penggunaan:*\n" +
			"1. Lihat daftar jadibot dengan `.listjadibot`\n" +
			"2. Copy ID jadibot yang ingin dihentikan\n" +
			"3. Kirim `.stopjadibot <id>`\n\n" +
			"*📝 Contoh:*\n" +
			"• `.stopjadibot abc123-def456-ghi789`\n\n" +
			"*⚠️ Catatan:*\n" +
			"• Jadibot yang dihentikan bisa diresume kembali\n" +
			"• Session tidak dihapus saat stop\n" +
			"• Untuk menghapus permanen, hubungi admin"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Cek apakah SessionManager tersedia
	if ctx.JadibotSessionManager == nil {
		message := "❌ *Fitur Jadibot belum diaktifkan!*\n\n" +
			"_SessionManager belum diinisialisasi. Hubungi admin bot._"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Ambil ID jadibot
	jadibotID := ctx.Args[0]

	// Cek apakah jadibot milik user
	botInfo, err := ctx.JadibotSessionManager.GetJadibotInfo(jadibotID)
	if err != nil {
		message := "❌ *Jadibot tidak ditemukan!*\n\n" +
			"┌─⦿ *Error*\n" +
			fmt.Sprintf("│ • ID: %s\n", jadibotID) +
			"│ • Jadibot mungkin sudah dihapus\n" +
			"└──────────────\n\n" +
			"*📝 Solusi:*\n" +
			"• Cek ID dengan `.listjadibot`\n" +
			"• Pastikan ID yang dimasukkan benar"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Verifikasi ownership
	if botInfo.OwnerJID != ctx.Sender.String() {
		message := "❌ *Akses Ditolak!*\n\n" +
			"┌─⦿ *Error*\n" +
			"│ • Jadibot ini bukan milik Anda\n" +
			"└──────────────\n\n" +
			"*📝 Solusi:*\n" +
			"• Gunakan `.listjadibot` untuk melihat jadibot Anda\n" +
			"• Hanya bisa mengontrol jadibot milik sendiri"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Cek apakah jadibot sedang running
	isRunning := ctx.JadibotSessionManager.IsRunning(jadibotID)
	if !isRunning {
		message := "ℹ️ *Jadibot Tidak Sedang Berjalan*\n\n" +
			"┌─⦿ *Info*\n" +
			fmt.Sprintf("│ • ID: %s\n", jadibotID) +
			fmt.Sprintf("│ • Status: %s\n", formatJadibotStatus(botInfo.Status)) +
			"└──────────────\n\n" +
			"*📝 Command yang tersedia:*\n" +
			"• `.resumejadibot <id>` - Resume jadibot\n" +
			"• `.listjadibot` - Lihat status jadibot"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Stop jadibot
	if err := ctx.JadibotSessionManager.StopJadibot(jadibotID); err != nil {
		message := "❌ *Gagal menghentikan jadibot!*\n\n" +
			"┌─⦿ *Error*\n" +
			fmt.Sprintf("│ • %v\n", err) +
			"└──────────────"
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Sukses
	message := "✅ *Jadibot Berhasil Dihentikan!*\n\n" +
		"┌─⦿ *Info*\n" +
		fmt.Sprintf("│ • ID: `%s`\n", jadibotID) +
		fmt.Sprintf("│ • Nomor: %s\n", botInfo.PhoneNumber) +
		"│ • Status: ⏹️ Stopped\n" +
		"└──────────────\n\n" +
		"*📝 Command yang tersedia:*\n" +
		fmt.Sprintf("• `.resumejadibot %s` - Resume jadibot\n", jadibotID) +
		"• `.listjadibot` - Lihat status jadibot\n\n" +
		"*💡 Tips:*\n" +
		"• Session jadibot tetap tersimpan\n" +
		"• Anda bisa resume kapan saja\n" +
		"• Jadibot tidak menggunakan resource saat stop"
	_, err = ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return err
}
