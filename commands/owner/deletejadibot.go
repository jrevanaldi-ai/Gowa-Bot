package owner

import (
	"fmt"
	"reflect"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/jrevanaldi-ai/gowa/proto/waE2E"
	"github.com/jrevanaldi-ai/gowa/types"

	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

var DeleteJadibotMetadata = &lib.CommandMetadata{
	Cmd:       "deletejadibot",
	Tag:       "owner",
	Desc:      "Hapus jadibot milik sendiri (stop + hapus session)",
	Example:   ".deletejadibot atau .deletejadibot <id>",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"deljadibot", "deljb", "deletejb", "delsession"},
}

func DeleteJadibotHandler(ctx *lib.CommandContext) error {
	if ctx.JadibotSessionManager == nil {
		message := "Fitur Jadibot belum diaktifkan.\n\n" +
			"SessionManager belum diinisialisasi. Hubungi admin bot."
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	myBots, err := ctx.JadibotSessionManager.GetJadibotByOwner(ctx.Sender.String())
	if err != nil {
		message := "Gagal mengambil data jadibot.\n\n" +
			fmt.Sprintf("Error: %v", err)
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	if len(myBots) == 0 {
		message := "Anda belum memiliki jadibot.\n\n" +
			"Buat jadibot dulu dengan:\n" +
			"- .jadibot <nomor_telepon>"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	var jadibotID string
	var confirm bool

	switch {
	case len(ctx.Args) == 0:
		if len(myBots) == 1 {
			jadibotID = myBots[0].ID
		} else {
			message := "Anda memiliki beberapa jadibot. Pilih salah satu untuk dihapus.\n\n" +
				"Daftar Jadibot Anda:\n"
			for i, bot := range myBots {
				message += fmt.Sprintf("%d. %s (%s)\n", i+1, bot.PhoneNumber, formatJadibotStatus(bot.Status)) +
					fmt.Sprintf("   .deletejadibot %s\n\n", bot.ID)
			}
			message += "Catatan:\n" +
				"- Tindakan ini akan menghapus session secara permanen"
			_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
			return err
		}
	case len(ctx.Args) >= 1:
		jadibotID = ctx.Args[0]
		if len(ctx.Args) > 1 && ctx.Args[1] == "--confirm" {
			confirm = true
		}
	}

	botInfo, err := ctx.JadibotSessionManager.GetJadibotInfo(jadibotID)
	if err != nil {
		message := "Jadibot tidak ditemukan.\n\n" +
			fmt.Sprintf("- ID: %s\n", jadibotID) +
			"- Mungkin sudah dihapus atau ID salah\n\n" +
			"Cek ID dengan .listjadibot"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	if !isJadibotOwner(botInfo.OwnerJID, ctx.Sender) && !ctx.IsOwner {
		message := "Anda tidak berhak menghapus jadibot ini.\n\n" +
			"- Hanya nomor yang membuat jadibot yang bisa menghapusnya\n" +
			"- Atau owner bot induk"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	if !confirm {
		isRunning := ctx.JadibotSessionManager.IsRunning(jadibotID)
		message := "KONFIRMASI HAPUS JADIBOT\n\n" +
			"Detail:\n" +
			fmt.Sprintf("- ID: %s\n", botInfo.ID) +
			fmt.Sprintf("- Nomor: %s\n", botInfo.PhoneNumber) +
			fmt.Sprintf("- Status: %s\n", formatJadibotStatus(botInfo.Status)) +
			fmt.Sprintf("- Running: %s\n\n", formatBoolean(isRunning)) +
			"Aksi yang akan dilakukan:\n" +
			"1. Stop jadibot (jika running)\n" +
			"2. Hapus session folder dari server\n" +
			"3. Hapus data dari database\n\n" +
			"PERINGATAN:\n" +
			"- Tindakan ini TIDAK DAPAT DIBATALKAN\n" +
			"- Session akan hilang permanen\n" +
			"- Anda harus pairing ulang jika ingin buat baru\n\n" +
			"Untuk konfirmasi, ketik:\n" +
			fmt.Sprintf(".deletejadibot %s --confirm", jadibotID)

		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	return executeDeleteJadibot(ctx, jadibotID, botInfo)
}

func executeDeleteJadibot(ctx *lib.CommandContext, jadibotID string, botInfo *lib.JadibotInfo) error {
	loadingMsg := "Menghapus jadibot...\n\n" +
		fmt.Sprintf("ID: %s\n", jadibotID) +
		"Status: Memulai...\n\n" +
		"Mohon tunggu..."

	sentResp, err := ctx.SendMessage(helper.CreateSimpleReply(loadingMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	if err != nil {
		return fmt.Errorf("failed to send loading message: %w", err)
	}

	var sentMsgID string
	respValue := reflect.ValueOf(sentResp)
	if respValue.Kind() == reflect.Struct {
		if idField := respValue.FieldByName("ID"); idField.IsValid() {
			sentMsgID = idField.String()
		}
	}

	editMessage := func(content string) {
		if sentMsgID == "" {
			_, _ = ctx.SendMessage(helper.CreateSimpleReply(content, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
			return
		}
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

	editMessage("Menghapus jadibot...\n\n" +
		fmt.Sprintf("ID: %s\n", jadibotID) +
		"Step 1/3: Menghentikan jadibot...\n\n" +
		"Mohon tunggu...")

	if ctx.JadibotSessionManager.IsRunning(jadibotID) {
		if err := ctx.JadibotSessionManager.StopJadibot(jadibotID); err != nil {
			editMessage("GAGAL MENGHAPUS JADIBOT.\n\n" +
				fmt.Sprintf("Gagal menghentikan jadibot: %v", err))
			return err
		}
	}

	editMessage("Menghapus jadibot...\n\n" +
		fmt.Sprintf("ID: %s\n", jadibotID) +
		"Step 2/3: Menghapus session folder...\n\n" +
		"Mohon tunggu...")

	time.Sleep(500 * time.Millisecond)

	editMessage("Menghapus jadibot...\n\n" +
		fmt.Sprintf("ID: %s\n", jadibotID) +
		"Step 3/3: Menghapus dari database...\n\n" +
		"Mohon tunggu...")

	if err := ctx.JadibotSessionManager.DeleteJadibot(jadibotID); err != nil {
		editMessage("GAGAL MENGHAPUS JADIBOT.\n\n" +
			fmt.Sprintf("Error: %v\n\n", err) +
			"Jadibot sudah dihentikan tapi data masih ada di database.")
		return err
	}

	editMessage("Jadibot berhasil dihapus.\n\n" +
		"Detail:\n" +
		fmt.Sprintf("- ID: %s\n", jadibotID) +
		fmt.Sprintf("- Nomor: %s\n\n", botInfo.PhoneNumber) +
		"Aksi selesai:\n" +
		"- Jadibot dihentikan\n" +
		"- Session folder dihapus\n" +
		"- Data database dihapus\n\n" +
		"Untuk membuat jadibot baru:\n" +
		"- .jadibot <nomor>")

	return nil
}

func isJadibotOwner(ownerJID string, sender types.JID) bool {
	owner := lib.StringToJID(ownerJID)
	return owner.User != "" && owner.User == sender.User
}

func formatJadibotStatus(status string) string {
	switch status {
	case "active":
		return "Aktif"
	case "paused":
		return "Paused"
	case "stopped":
		return "Berhenti"
	default:
		return "Tidak Diketahui"
	}
}

func formatBoolean(b bool) string {
	if b {
		return "Ya"
	}
	return "Tidak"
}
