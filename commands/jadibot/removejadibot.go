package jadibot

import (
	"fmt"
	"reflect"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/jrevanaldi-ai/gowa/proto/waE2E"
	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)


var RemoveJadibotMetadata = &lib.CommandMetadata{
	Cmd:       "removejadibot",
	Tag:       "jadibot",
	Desc:      "Hapus jadibot secara permanen (owner only)",
	Example:   ".removejadibot <id_jadibot>",
	Hidden:    false,
	OwnerOnly: true,
	Alias:     []string{"removejb", "hapusjadibot", "hapusjb"},
}


func RemoveJadibotHandler(ctx *lib.CommandContext) error {

	if len(ctx.Args) == 0 {
		message := "*🗑️ Remove Jadibot (Owner Only)*\n\n" +
			"┌─⦿ *Usage*\n" +
			"│ • `.removejadibot <id_jadibot>` - Hapus jadibot permanen\n" +
			"└──────────────\n\n" +
			"*📋 Cara Penggunaan:*\n" +
			"1. Lihat daftar semua jadibot dengan `.listjadibot`\n" +
			"2. Copy ID jadibot yang ingin dihapus\n" +
			"3. Kirim `.removejadibot <id>`\n\n" +
			"*📝 Contoh:*\n" +
			"• `.removejadibot abc123-def456-ghi789`\n\n" +
			"*⚠️ PERINGATAN:*\n" +
			"• Command ini HANYA untuk owner bot induk\n" +
			"• Jadibot akan dihapus SECARA PERMANEN\n" +
			"• Session akan dihapus dari server\n" +
			"• Data tidak bisa dikembalikan\n" +
			"• User harus buat jadibot baru dari awal"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	if ctx.JadibotSessionManager == nil {
		message := "❌ *Fitur Jadibot belum diaktifkan!*\n\n" +
			"_SessionManager belum diinisialisasi. Hubungi admin bot._"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	jadibotID := ctx.Args[0]


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


	if len(ctx.Args) > 1 && ctx.Args[1] == "--confirm" {

		return executeRemoveJadibotWithEdit(ctx, jadibotID, botInfo)
	}


	isRunning := ctx.JadibotSessionManager.IsRunning(jadibotID)


	message := "*⚠️ KONFIRMASI HAPUS JADIBOT*\n\n" +
		"┌─⦿ *Detail Jadibot*\n" +
		fmt.Sprintf("│ • ID: `%s`\n", botInfo.ID) +
		fmt.Sprintf("│ • Pemilik: %s\n", botInfo.OwnerJID) +
		fmt.Sprintf("│ • Nomor: %s\n", botInfo.PhoneNumber) +
		fmt.Sprintf("│ • Status: %s\n", formatJadibotStatus(botInfo.Status)) +
		fmt.Sprintf("│ • Sedang Running: %s\n", formatBoolean(isRunning)) +
		"└──────────────\n\n" +
		"*🗑️ Aksi yang akan dilakukan:*\n" +
		"1. ⏹️ Stop jadibot (jika sedang running)\n" +
		"2. 📁 Hapus session folder dari server\n" +
		"3. 🗄️ Hapus data dari database\n" +
		"4. ❌ Jadibot tidak bisa digunakan lagi\n\n" +
		"*⚠️ PERINGATAN:*\n" +
		"• Tindakan ini TIDAK DAPAT DIBATALKAN\n" +
		"• Semua data jadibot akan hilang permanen\n" +
		"• User harus pairing ulang jika ingin membuat baru\n\n" +
		"*📝 Untuk konfirmasi, ketik:*\n" +
		fmt.Sprintf("`.removejadibot %s --confirm`", jadibotID)

	_, err = ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return err
}


func executeRemoveJadibotWithEdit(ctx *lib.CommandContext, jadibotID string, botInfo *lib.JadibotInfo) error {

	loadingMsg := "🔄 *Menghapus jadibot...*\n\n" +
		"┌─⦿ *Progress*\n" +
		fmt.Sprintf("│ • ID: `%s`\n", jadibotID) +
		"│ • Status: ⏳ Memulai...\n" +
		"└──────────────\n\n" +
		"_Mohon tunggu..._"

	sentResp, err := ctx.SendMessage(helper.CreateSimpleReply(loadingMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	if err != nil {
		return fmt.Errorf("failed to send loading message: %w", err)
	}


	var sentMsgID string
	respValue := reflect.ValueOf(sentResp)
	if respValue.Kind() == reflect.Struct {
		idField := respValue.FieldByName("ID")
		if idField.IsValid() {
			sentMsgID = idField.String()
		}
	}


	if sentMsgID == "" {
		return executeRemoveJadibotFallback(ctx, jadibotID, botInfo)
	}


	editMessage := func(content string) {
		editMsg := ctx.Client.BuildEdit(ctx.Chat, sentMsgID, &waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text: proto.String(content),
				ContextInfo: &waE2E.ContextInfo{
					StanzaID:    proto.String(ctx.MessageID),
					Participant: proto.String(ctx.Sender.String()),
				},
			},
		})
		_, _ = ctx.Client.SendMessage(ctx.Ctx, ctx.Chat, editMsg)

		time.Sleep(500 * time.Millisecond)
	}


	editMessage("🔄 *Menghapus jadibot...*\n\n" +
		"┌─⦿ *Progress*\n" +
		fmt.Sprintf("│ • ID: `%s`\n", jadibotID) +
		"│ • Status: ⏹️ Menghentikan jadibot...\n" +
		"│ • Step: 1/3\n" +
		"└──────────────\n\n" +
		"_Mohon tunggu..._")

	if ctx.JadibotSessionManager.IsRunning(jadibotID) {
		if err := ctx.JadibotSessionManager.StopJadibot(jadibotID); err != nil {
			editMessage("❌ *GAGAL MENGHAPUS JADIBOT!*\n\n" +
				"┌─⦿ *Error*\n" +
				fmt.Sprintf("│ • Gagal menghentikan jadibot: %v\n", err) +
				"└──────────────\n\n" +
				"*📝 Catatan:*\n" +
				"• Jadibot mungkin sudah terhenti\n" +
				"• Coba lagi atau hubungi admin")
			return err
		}
	}


	editMessage("🔄 *Menghapus jadibot...*\n\n" +
		"┌─⦿ *Progress*\n" +
		fmt.Sprintf("│ • ID: `%s`\n", jadibotID) +
		"│ • Status: ✅ Jadibot dihentikan\n" +
		"│ • Step: 2/3 - 📁 Menghapus session folder...\n" +
		"└──────────────\n\n" +
		"_Mohon tunggu..._")


	time.Sleep(1 * time.Second)


	editMessage("🔄 *Menghapus jadibot...*\n\n" +
		"┌─⦿ *Progress*\n" +
		fmt.Sprintf("│ • ID: `%s`\n", jadibotID) +
		"│ • Status: ✅ Jadibot dihentikan\n" +
		"│ • Step: 3/3 - 🗄️ Menghapus dari database...\n" +
		"└──────────────\n\n" +
		"_Mohon tunggu..._")

	if err := ctx.JadibotSessionManager.DeleteJadibot(jadibotID); err != nil {
		editMessage("❌ *GAGAL MENGHAPUS JADIBOT!*\n\n" +
			"┌─⦿ *Error*\n" +
			fmt.Sprintf("│ • %v\n", err) +
			"└──────────────\n\n" +
			"*📝 Catatan:*\n" +
			"• Jadibot sudah dihentikan\n" +
			"• Tapi data masih ada di database\n" +
			"• Coba lagi atau hubungi admin")
		return err
	}


	editMessage("✅ *JADIBOT BERHASIL DIHAPUS!*\n\n" +
		"┌─⦿ *Detail*\n" +
		fmt.Sprintf("│ • ID: `%s`\n", jadibotID) +
		fmt.Sprintf("│ • Pemilik: %s\n", botInfo.OwnerJID) +
		fmt.Sprintf("│ • Nomor: %s\n", botInfo.PhoneNumber) +
		"└──────────────\n\n" +
		"┌─⦿ *Progress*\n" +
		"│ • ✅ Step 1/3: Jadibot dihentikan\n" +
		"│ • ✅ Step 2/3: Session folder dihapus\n" +
		"│ • ✅ Step 3/3: Data database dihapus\n" +
		"└──────────────\n\n" +
		"*🗑️ Aksi telah selesai:*\n" +
		"• Jadibot sudah tidak bisa digunakan\n" +
		"• Session dihapus dari server\n" +
		"• Data dihapus dari database\n\n" +
		"*📝 Catatan:*\n" +
		"• User harus pairing ulang untuk buat baru\n" +
		"• User bisa buat jadibot dengan `.jadibot`")

	return nil
}


func executeRemoveJadibotFallback(ctx *lib.CommandContext, jadibotID string, botInfo *lib.JadibotInfo) error {

	loadingMsg := "🔄 *Menghapus jadibot...*\n\n" +
		"┌─⦿ *Progress*\n" +
		fmt.Sprintf("│ • ID: %s\n", jadibotID) +
		"│ • Step 1/3: Menghentikan jadibot...\n" +
		"└──────────────\n\n" +
		"_Mohon tunggu..._"
	_, _ = ctx.SendMessage(helper.CreateSimpleReply(loadingMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))


	if ctx.JadibotSessionManager.IsRunning(jadibotID) {
		if err := ctx.JadibotSessionManager.StopJadibot(jadibotID); err != nil {
			return fmt.Errorf("failed to stop jadibot: %w", err)
		}
	}

	time.Sleep(500 * time.Millisecond)


	if err := ctx.JadibotSessionManager.DeleteJadibot(jadibotID); err != nil {
		return fmt.Errorf("failed to delete jadibot: %w", err)
	}


	message := "✅ *Jadibot Berhasil Dihapus Secara Permanen!*\n\n" +
		"┌─⦿ *Detail*\n" +
		fmt.Sprintf("│ • ID: `%s`\n", jadibotID) +
		fmt.Sprintf("│ • Pemilik: %s\n", botInfo.OwnerJID) +
		fmt.Sprintf("│ • Nomor: %s\n", botInfo.PhoneNumber) +
		"└──────────────\n\n" +
		"*🗑️ Aksi yang telah dilakukan:*\n" +
		"1. ✅ Jadibot dihentikan\n" +
		"2. ✅ Session folder dihapus\n" +
		"3. ✅ Data dihapus dari database\n\n" +
		"*📝 Catatan:*\n" +
		"• Jadibot sudah tidak bisa digunakan lagi\n" +
		"• Jika user ingin membuat baru, harus pairing ulang\n" +
		"• User bisa membuat jadibot baru dengan `.jadibot`"

	_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return err
}


func formatBoolean(b bool) string {
	if b {
		return "✅ Ya"
	}
	return "❌ Tidak"
}
