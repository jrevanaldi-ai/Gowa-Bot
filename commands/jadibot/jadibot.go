package jadibot

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

// JadibotMetadata adalah metadata untuk command jadibot
var JadibotMetadata = &lib.CommandMetadata{
	Cmd:       "jadibot",
	Tag:       "jadibot",
	Desc:      "Buat bot WhatsApp baru melalui bot induk (dapat pairing code)",
	Example:   ".jadibot 6281234567890",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"jb", "botbaru"},
}

// JadibotHandler menangani command jadibot
func JadibotHandler(ctx *lib.CommandContext) error {
	// Cek apakah ada argument (nomor telepon)
	if len(ctx.Args) == 0 {
		message := "*🤖 Jadibot - Buat Bot Baru*\n\n" +
			"Buat bot WhatsApp pribadi Anda melalui bot induk!\n\n" +
			"┌─⦿ *Usage*\n" +
			"│ • `.jadibot <nomor_telepon>` - Buat jadibot baru\n" +
			"└──────────────\n\n" +
			"*📋 Contoh:*\n" +
			"• `.jadibot 6281234567890`\n\n" +
			"*📝 Cara Penggunaan:*\n" +
			"1. Kirim command dengan nomor Anda\n" +
			"2. Bot akan memberikan pairing code\n" +
			"3. Buka WhatsApp → Perangkat Tertaut → Tautkan\n" +
			"4. Masukkan pairing code yang diberikan\n" +
			"5. ✅ Bot Anda siap digunakan!\n\n" +
			"*⚠️ Catatan:*\n" +
			"• Gunakan format internasional (62xxx)\n" +
			"• Tanpa tanda + atau spasi\n" +
			"• Satu nomor = satu bot\n" +
			"• Bot memiliki fitur yang sama dengan bot induk"
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

	// Ambil nomor telepon dari args
	phoneNumber := strings.Join(ctx.Args, " ")
	
	// Validasi format nomor telepon
	phoneNumber = strings.TrimSpace(phoneNumber)
	phoneNumber = strings.TrimPrefix(phoneNumber, "+")
	
	// Cek apakah nomor valid
	if !isValidPhoneNumber(phoneNumber) {
		message := "❌ *Nomor telepon tidak valid!*\n\n" +
			"┌─⦿ *Format yang benar*\n" +
			"│ • Gunakan format internasional\n" +
			"│ • Tanpa tanda + atau spasi\n" +
			"│ • Contoh: 6281234567890\n" +
			"└──────────────\n\n" +
			"*📝 Contoh:*\n" +
			"• `.jadibot 6281234567890` ✅\n" +
			"• `.jadibot +62 812-3456-7890` ❌"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Cek apakah owner (owner bisa unlimited jadibot)
	isOwner := ctx.IsOwner

	// Jika bukan owner, cek apakah user sudah punya jadibot (limit 1 untuk non-owner)
	if !isOwner {
		existingBots, err := ctx.JadibotSessionManager.GetJadibotByOwner(ctx.Sender.String())
		if err == nil && len(existingBots) > 0 {
			message := "❌ *Anda sudah membuat jadibot!*\n\n" +
				fmt.Sprintf("┌─⦿ *Jadibot Anda*\n│ • Jumlah: %d bot\n│ • ID: %s\n└──────────────\n\n", len(existingBots), existingBots[0].ID) +
				"*📋 Command yang tersedia:*\n" +
				"• `.listjadibot` - Lihat status jadibot\n" +
				"• `.stopjadibot <id>` - Hentikan jadibot\n" +
				"• `.pausejadibot <id>` - Pause jadibot\n" +
				"• `.resumejadibot <id>` - Resume jadibot\n\n" +
				"*⚠️ Catatan:*\n" +
				"• Hanya owner yang bisa membuat jadibot unlimited"
			_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
			return err
		}
	}

	// Kirim pesan loading
	loadingMsg := "🔄 *Memproses pembuatan jadibot...*\n\n" +
		"┌─⦿ *Info*\n" +
		fmt.Sprintf("│ • Nomor: %s\n", phoneNumber) +
		"│ • Status: Membuat session...\n" +
		"└──────────────\n\n" +
		"_Mohon tunggu sebentar..._"
	_, sendErr := ctx.SendMessage(helper.CreateSimpleReply(loadingMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	if sendErr != nil {
		return fmt.Errorf("failed to send loading message: %w", sendErr)
	}

	// Create jadibot
	jadibotID, createErr := ctx.JadibotSessionManager.CreateJadibot(ctx.Ctx, ctx.Sender.String(), phoneNumber)
	if createErr != nil {
		errorMsg := fmt.Sprintf("❌ *Gagal membuat jadibot!*\n\n┌─⦿ *Error*\n│ • %v\n└──────────────", createErr)
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return createErr
	}

	// Start jadibot (akan generate pairing code)
	pairingCode, startErr := ctx.JadibotSessionManager.StartJadibot(ctx.Ctx, jadibotID, phoneNumber)
	if startErr != nil {
		errorMsg := fmt.Sprintf("❌ *Gagal memulai jadibot!*\n\n┌─⦿ *Error*\n│ • %v\n└──────────────", startErr)
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return startErr
	}

	// Jika pairing code tidak kosong (belum paired), kirim ke user
	if pairingCode != "" {
		message := "✅ *Jadibot Berhasil Dibuat!*\n\n" +
			"┌─⦿ *Info Jadibot*\n" +
			fmt.Sprintf("│ • ID: `%s`\n", jadibotID) +
			fmt.Sprintf("│ • Nomor: %s\n", phoneNumber) +
			fmt.Sprintf("│ • Pairing Code: `%s`\n", pairingCode) +
			"└──────────────\n\n" +
			"*📝 Cara Pairing:*\n" +
			"1. Buka WhatsApp di HP Anda\n" +
			"2. Menu → Perangkat Tertaut\n" +
			"3. Tautkan Perangkat\n" +
			"4. Masukkan pairing code di atas\n\n" +
			"*⚠️ Penting:*\n" +
			"• Pairing code kadaluarsa dalam 160 detik\n" +
			"• Gunakan jadibot dengan bijak\n" +
			"• Bot memiliki fitur yang sama dengan bot induk\n\n" +
			"*📖 Command Jadibot:*\n" +
			"• `.listjadibot` - Cek status\n" +
			"• `.stopjadibot <id>` - Stop bot\n" +
			"• `.pausejadibot <id>` - Pause bot"
		_, sendErr := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return sendErr
	}

	// Jika pairing code kosong, berarti sudah paired
	message := "✅ *Jadibot Sudah Aktif!*\n\n" +
		"┌─⦿ *Info Jadibot*\n" +
		fmt.Sprintf("│ • ID: `%s`\n", jadibotID) +
		fmt.Sprintf("│ • Nomor: %s\n", phoneNumber) +
		"│ • Status: Sudah paired ✓\n" +
		"└──────────────\n\n" +
		"Jadibot Anda sudah aktif dan siap digunakan!\n\n" +
		"*📖 Command Jadibot:*\n" +
		"• `.listjadibot` - Cek status\n" +
		"• `.stopjadibot <id>` - Stop bot\n" +
		"• `.pausejadibot <id>` - Pause bot"
	_, sendErr2 := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return sendErr2
}

// isValidPhoneNumber validasi format nomor telepon
func isValidPhoneNumber(phone string) bool {
	// Harus diawali dengan angka dan panjang 10-15 digit
	re := regexp.MustCompile(`^[0-9]{10,15}$`)
	return re.MatchString(phone)
}
