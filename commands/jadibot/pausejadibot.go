package jadibot

import (
	"fmt"

	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

// PauseJadibotMetadata adalah metadata untuk command pausejadibot
var PauseJadibotMetadata = &lib.CommandMetadata{
	Cmd:       "pausejadibot",
	Tag:       "jadibot",
	Desc:      "Pause jadibot yang sedang berjalan (bisa diresume)",
	Example:   ".pausejadibot <id_jadibot>",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"pausejb", "jeda"},
}

// PauseJadibotHandler menangani command pausejadibot
func PauseJadibotHandler(ctx *lib.CommandContext) error {
	// Cek apakah ada argument (ID jadibot)
	if len(ctx.Args) == 0 {
		message := "*⏸️ Pause Jadibot*\n\n" +
			"┌─⦿ *Usage*\n" +
			"│ • `.pausejadibot <id_jadibot>` - Pause jadibot\n" +
			"└──────────────\n\n" +
			"*📋 Cara Penggunaan:*\n" +
			"1. Lihat daftar jadibot dengan `.listjadibot`\n" +
			"2. Copy ID jadibot yang ingin di-pause\n" +
			"3. Kirim `.pausejadibot <id>`\n\n" +
			"*📝 Contoh:*\n" +
			"• `.pausejadibot abc123-def456-ghi789`\n\n" +
			"*⚠️ Catatan:*\n" +
			"• Jadibot yang di-pause bisa diresume kembali\n" +
			"• Session tidak dihapus saat pause\n" +
			"• Gunakan `.resumejadibot` untuk melanjutkan"
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
			fmt.Sprintf("• `.resumejadibot %s` - Resume jadibot\n", jadibotID) +
			"• `.listjadibot` - Lihat status jadibot"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Pause jadibot
	if err := ctx.JadibotSessionManager.PauseJadibot(jadibotID); err != nil {
		message := "❌ *Gagal pause jadibot!*\n\n" +
			"┌─⦿ *Error*\n" +
			fmt.Sprintf("│ • %v\n", err) +
			"└──────────────"
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Sukses
	message := "✅ *Jadibot Berhasil Di-Pause!*\n\n" +
		"┌─⦿ *Info*\n" +
		fmt.Sprintf("│ • ID: `%s`\n", jadibotID) +
		fmt.Sprintf("│ • Nomor: %s\n", botInfo.PhoneNumber) +
		"│ • Status: ⏸️ Paused\n" +
		"└──────────────\n\n" +
		"*📝 Command yang tersedia:*\n" +
		fmt.Sprintf("• `.resumejadibot %s` - Resume jadibot\n", jadibotID) +
		"• `.listjadibot` - Lihat status jadibot\n\n" +
		"*💡 Tips:*\n" +
		"• Jadibot pause tidak menggunakan resource\n" +
		"• Session tetap tersimpan\n" +
		"• Resume kapan saja saat dibutuhkan"
	_, err = ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return err
}

// ResumeJadibotMetadata adalah metadata untuk command resumejadibot
var ResumeJadibotMetadata = &lib.CommandMetadata{
	Cmd:       "resumejadibot",
	Tag:       "jadibot",
	Desc:      "Resume jadibot yang sedang di-pause atau stopped",
	Example:   ".resumejadibot <id_jadibot>",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"resumejb", "lanjut"},
}

// ResumeJadibotHandler menangani command resumejadibot
func ResumeJadibotHandler(ctx *lib.CommandContext) error {
	// Cek apakah ada argument (ID jadibot)
	if len(ctx.Args) == 0 {
		message := "*▶️ Resume Jadibot*\n\n" +
			"┌─⦿ *Usage*\n" +
			"│ • `.resumejadibot <id_jadibot>` - Resume jadibot\n" +
			"└──────────────\n\n" +
			"*📋 Cara Penggunaan:*\n" +
			"1. Lihat daftar jadibot dengan `.listjadibot`\n" +
			"2. Copy ID jadibot yang ingin di-resume\n" +
			"3. Kirim `.resumejadibot <id>`\n\n" +
			"*📝 Contoh:*\n" +
			"• `.resumejadibot abc123-def456-ghi789`\n\n" +
			"*⚠️ Catatan:*\n" +
			"• Jadibot yang di-resume akan aktif kembali\n" +
			"• Session akan digunakan kembali\n" +
			"• Tidak perlu pairing ulang"
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

	// Cek apakah jadibot sudah running
	isRunning := ctx.JadibotSessionManager.IsRunning(jadibotID)
	if isRunning {
		message := "ℹ️ *Jadibot Sudah Berjalan*\n\n" +
			"┌─⦿ *Info*\n" +
			fmt.Sprintf("│ • ID: %s\n", jadibotID) +
			fmt.Sprintf("│ • Status: %s\n", formatJadibotStatus(botInfo.Status)) +
			"└──────────────\n\n" +
			"*📝 Command yang tersedia:*\n" +
			fmt.Sprintf("• `.stopjadibot %s` - Stop jadibot\n", jadibotID) +
			fmt.Sprintf("• `.pausejadibot %s` - Pause jadibot\n", jadibotID) +
			"• `.listjadibot` - Lihat status jadibot"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Kirim pesan loading
	loadingMsg := "🔄 *Memproses resume jadibot...*\n\n" +
		"┌─⦿ *Info*\n" +
		fmt.Sprintf("│ • ID: %s\n", jadibotID) +
		"│ • Status: Memulai kembali...\n" +
		"└──────────────\n\n" +
		"_Mohon tunggu sebentar..._"
	_, err = ctx.SendMessage(helper.CreateSimpleReply(loadingMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	if err != nil {
		return fmt.Errorf("failed to send loading message: %w", err)
	}

	// Resume jadibot
	if err := ctx.JadibotSessionManager.ResumeJadibot(ctx.Ctx, jadibotID); err != nil {
		message := "❌ *Gagal resume jadibot!*\n\n" +
			"┌─⦿ *Error*\n" +
			fmt.Sprintf("│ • %v\n", err) +
			"└──────────────\n\n" +
			"*📝 Solusi:*\n" +
			"• Coba lagi dalam beberapa saat\n" +
			"• Jika masih error, buat jadibot baru dengan `.jadibot`"
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Sukses
	message := "✅ *Jadibot Berhasil Di-Resume!*\n\n" +
		"┌─⦿ *Info*\n" +
		fmt.Sprintf("│ • ID: `%s`\n", jadibotID) +
		fmt.Sprintf("│ • Nomor: %s\n", botInfo.PhoneNumber) +
		"│ • Status: 🟢 Aktif\n" +
		"└──────────────\n\n" +
		"Jadibot Anda sudah aktif kembali!\n\n" +
		"*📝 Command yang tersedia:*\n" +
		fmt.Sprintf("• `.stopjadibot %s` - Stop jadibot\n", jadibotID) +
		fmt.Sprintf("• `.pausejadibot %s` - Pause jadibot\n", jadibotID) +
		"• `.listjadibot` - Lihat status jadibot"
	_, err = ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return err
}
