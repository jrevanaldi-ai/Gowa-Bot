package general

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/google/uuid"
	"github.com/jrevanaldi-ai/gowa"
	"github.com/jrevanaldi-ai/gowa/proto/waE2E"
	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)


var DonasiMetadata = &lib.CommandMetadata{
	Cmd:       "donasi",
	Tag:       "main",
	Desc:      "Donasi ke bot via QRIS MustikaPay",
	Example:   ".donasi 10000",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"donate", "qr"},
}


func DonasiHandler(ctx *lib.CommandContext) error {

	if len(ctx.Args) == 0 {
		message := "*💰 Donasi Gowa-Bot*\n\n" +
			"┌─⦿ *Usage*\n" +
			"│ • `.donasi <nominal>` - Buat QRIS payment\n" +
			"└──────────────\n\n" +
			"*📝 Contoh:*\n" +
			"• `.donasi 10000` - Donasi Rp 10.000\n" +
			"• `.donasi 50000` - Donasi Rp 50.000\n" +
			"• `.donasi 100000` - Donasi Rp 100.000\n\n" +
			"*💡 Info:*\n" +
			"• Minimal donasi: Rp 10\n" +
			"• Maksimal donasi: Rp 10.000.000\n" +
			"• QRIS berlaku selama 2 menit\n" +
			"• Biaya MDR: 0.7%\n\n" +
			"*🔒 Pembayaran aman via MustikaPay*"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	amountStr := ctx.Args[0]
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		message := "❌ *Nominal tidak valid!*\n\n" +
			"┌─⦿ *Info*\n" +
			"│ • Nominal harus berupa angka\n" +
			"│ • Contoh: `.donasi 10000`\n" +
			"└──────────────"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	if amount < 10 {
		message := "❌ *Minimal donasi: Rp 10!*\n\n" +
			"┌─⦿ *Info*\n" +
			"│ • Nominal terlalu kecil\n" +
			"│ • Minimal: Rp 10\n" +
			"└──────────────"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	if amount > 10000000 {
		message := "❌ *Maksimal donasi: Rp 10.000.000!*\n\n" +
			"┌─⦿ *Info*\n" +
			"│ • Nominal terlalu besar\n" +
			"│ • Maksimal: Rp 10.000.000\n" +
			"└──────────────"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	dbManager, ok := ctx.BotClient.GetDBManager().(*helper.DatabaseManager)
	if !ok || dbManager == nil {
		return fmt.Errorf("database manager tidak tersedia")
	}


	apiKey := getMustikaPayAPIKey()
	if apiKey == "" {
		message := "❌ *Fitur donasi belum dikonfigurasi!*\n\n" +
			"┌─⦿ *Info*\n" +
			"│ • Hubungi admin untuk aktivasi\n" +
			"└──────────────"
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}


	loadingMsg := "⏳ *Membuat QRIS payment...*\n\n" +
		"┌─⦿ *Info*\n" +
		fmt.Sprintf("│ • Nominal: Rp %s\n", helper.FormatAmount(amount)) +
		"│ • Status: Memproses...\n" +
		"└──────────────\n\n" +
		"_Mohon tunggu..._"
	_, err = ctx.SendMessage(helper.CreateSimpleReply(loadingMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	if err != nil {
		return fmt.Errorf("failed to send loading message: %w", err)
	}


	logger := helper.NewLogger("Donasi")
	mustikaClient := helper.NewMustikaPayClient(apiKey, logger)

	logger.Info("Creating QRIS payment - Amount: %d, Customer: %s", amount, ctx.PushName)

	paymentReq := helper.CreatePaymentRequest{
		Amount:       amount,
		ProductName:  "Donasi Gowa-Bot",
		CustomerName: ctx.PushName,
	}

	paymentResp, err := mustikaClient.CreateQRIS(paymentReq)
	if err != nil {
		logger.Error("Failed to create QRIS: %v", err)
		errorMsg := fmt.Sprintf("❌ *Gagal membuat QRIS!*\n\n"+
			"┌─⦿ *Error*\n"+
			"│ • %s\n"+
			"└──────────────\n\n"+
			"*📝 Solusi:*\n"+
			"• Coba lagi dalam beberapa saat\n"+
			"• Hubungi admin jika masalah berlanjut\n"+
			"• Pastikan koneksi server stabil", err.Error())
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	logger.Info("QRIS created successfully - RefNo: %s, QR URL: %s", paymentResp.RefNo, paymentResp.GetQRImageURL())


	donationID := uuid.New().String()
	donationInfo := helper.DonationInfo{
		ID:           donationID,
		RefNo:        paymentResp.RefNo,
		UserJID:      ctx.Sender.String(),
		UserName:     ctx.PushName,
		Amount:       amount,
		QRString:     paymentResp.GetQRString(),
		QRImageURL:   paymentResp.GetQRImageURL(),
		Status:       "pending",
		ProductName:  paymentResp.ProductName,
	}

	if err := dbManager.CreateDonation(donationInfo); err != nil {
		logger.Warning("Failed to save donation to database: %v", err)

	}


	caption := fmt.Sprintf(
		"✅ *QRIS Payment Berhasil Dibuat!*\n\n"+
			"┌─⦿ *Detail Pembayaran*\n"+
			"│ • *Ref No:* `%s`\n"+
			"│ • *Nominal:* Rp %s\n"+
			"│ • *Produk:* %s\n"+
			"└──────────────\n\n"+
			"📱 *Cara Pembayaran:*\n"+
			"1. 📸 Scan QR Code di atas\n"+
			"2. 💰 Masukkan nominal: Rp %s\n"+
			"3. ✅ Konfirmasi pembayaran\n\n"+
			"⏰ *Batas Waktu:* 2 menit\n"+
			"🔍 *Cek Status:* `.cekdonasi %s`\n\n"+
			"_Bot akan otomatis mendeteksi pembayaran Anda!_",
		paymentResp.RefNo,
		helper.FormatAmount(amount),
		paymentResp.ProductName,
		helper.FormatAmount(amount),
		paymentResp.RefNo,
	)


	qrImageURL := paymentResp.GetQRImageURL()
	if qrImageURL == "" {

		errorMsg := "❌ *QR Image URL tidak tersedia!*\n\n" +
			"┌─⦿ *Info*\n" +
			"│ • QR URL kosong dari API\n" +
			"│ • Coba lagi dalam beberapa saat\n" +
			"└──────────────"
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}


	logger.Info("Downloading QR image from: %s", qrImageURL)
	httpClient := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := httpClient.Get(qrImageURL)
	if err != nil {
		logger.Error("Failed to download QR image: %v", err)
		errorMsg := fmt.Sprintf("❌ *Gagal download QR Image!*\n\n"+
			"┌─⦿ *Error*\n"+
			"│ • %s\n"+
			"└──────────────", err.Error())
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Failed to download QR image: HTTP %d", resp.StatusCode)
		errorMsg := fmt.Sprintf("❌ *Gagal download QR Image (HTTP %d)!*\n\n"+
			"┌─⦿ *Info*\n"+
			"│ • Server API error\n"+
			"└──────────────", resp.StatusCode)
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	qrImageData, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read QR image data: %v", err)
		errorMsg := fmt.Sprintf("❌ *Gagal membaca QR Image!*\n\n"+
			"┌─⦿ *Error*\n"+
			"│ • %s\n"+
			"└──────────────", err.Error())
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	logger.Info("QR image downloaded successfully (%d bytes)", len(qrImageData))


	uploadResp, err := ctx.Client.Upload(context.Background(), qrImageData, gowa.MediaImage)
	if err != nil {
		logger.Error("Failed to upload QR image: %v", err)
		errorMsg := fmt.Sprintf("❌ *Gagal upload QR Image!*\n\n"+
			"┌─⦿ *Error*\n"+
			"│ • %s\n"+
			"└──────────────", err.Error())
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	logger.Info("QR image uploaded successfully - URL: %s", uploadResp.URL)


	qrImageMsg := &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			URL:           proto.String(uploadResp.URL),
			DirectPath:    proto.String(uploadResp.DirectPath),
			Mimetype:      proto.String("image/png"),
			Caption:       proto.String(caption),
			FileSHA256:    uploadResp.FileSHA256,
			FileEncSHA256: uploadResp.FileEncSHA256,
			FileLength:    proto.Uint64(uploadResp.FileLength),
			MediaKey:      uploadResp.MediaKey,
			MediaKeyTimestamp: proto.Int64(time.Now().Unix()),
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:    proto.String(ctx.MessageID),
				Participant: proto.String(ctx.Sender.String()),
			},
		},
	}

	_, err = ctx.SendMessage(qrImageMsg)
	if err != nil {
		return fmt.Errorf("failed to send QR image: %w", err)
	}


	go monitorDonation(ctx, mustikaClient, dbManager, paymentResp.RefNo, ctx.Sender.String(), ctx.PushName, amount)

	return nil
}


func monitorDonation(ctx *lib.CommandContext, client *helper.MustikaPayClient, dbManager *helper.DatabaseManager, refNo string, userJID string, userName string, amount int) {
	logger := helper.NewLogger("DonasiMonitor")
	logger.Info("Start monitoring donation: %s", refNo)


	maxAttempts := 24
	interval := 5 * time.Second

	for i := 0; i < maxAttempts; i++ {
		time.Sleep(interval)


		status, err := client.CheckPaymentStatus(refNo)
		if err != nil {
			logger.Warning("Failed to check payment status (attempt %d/%d): %v", i+1, maxAttempts, err)
			continue
		}


		if helper.PaymentStatus(status.Status) == helper.StatusSuccess {

			dbManager.UpdateDonationStatus(refNo, helper.DonationSuccess, status.Type, status.Issuer, status.Payor)


			sendSuccessNotification(ctx, refNo, amount, status)
			return
		}


		if helper.PaymentStatus(status.Status) == helper.StatusExpired || helper.PaymentStatus(status.Status) == helper.StatusFailed {

			dbManager.UpdateDonationStatus(refNo, helper.DonationExpired, status.Type, status.Issuer, status.Payor)


			sendExpiredNotification(ctx, refNo, amount)
			return
		}


		if (i+1)%18 == 0 {
			logger.Debug("Payment still pending: %s (attempt %d/%d)", refNo, i+1, maxAttempts)
		}
	}


	dbManager.UpdateDonationStatus(refNo, helper.DonationExpired, "", "", "")
	sendTimeoutNotification(ctx, refNo, amount)
}


func sendSuccessNotification(ctx *lib.CommandContext, refNo string, amount int, status *helper.CheckPaymentResponse) {
	message := fmt.Sprintf(
		"🎉 *Pembayaran Berhasil!*\n\n"+
			"┌─⦿ *Detail*\n"+
			"│ • *Ref No:* `%s`\n"+
			"│ • *Nominal:* Rp %s\n"+
			"│ • *Metode:* %s\n"+
			"│ • *Pembayar:* %s\n"+
			"│ • *Status:* ✅ SUCCESS\n"+
			"└──────────────\n\n"+
			"_Terima kasih atas donasi Anda! 🙏_\n"+
			"_Dukungan Anda sangat berarti untuk pengembangan bot ini._",
		refNo,
		helper.FormatAmount(amount),
		status.Type,
		status.Payor,
	)

	_, _ = ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
}


func sendExpiredNotification(ctx *lib.CommandContext, refNo string, amount int) {
	message := fmt.Sprintf(
		"⏰ *QRIS Kadaluarsa!*\n\n"+
			"┌─⦿ *Detail*\n"+
			"│ • *Ref No:* `%s`\n"+
			"│ • *Nominal:* Rp %s\n"+
			"│ • *Status:* ❌ EXPIRED\n"+
			"└──────────────\n\n"+
			"*📝 Solusi:*\n"+
			"• Buat QRIS baru dengan `.donasi %d`\n"+
			"• Pastikan pembayaran dilakukan dalam 2 menit",
		refNo,
		helper.FormatAmount(amount),
		amount,
	)

	_, _ = ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
}


func sendTimeoutNotification(ctx *lib.CommandContext, refNo string, amount int) {
	message := fmt.Sprintf(
		"⏰ *Waktu Pembayaran Habis!*\n\n"+
			"┌─⦿ *Detail*\n"+
			"│ • *Ref No:* `%s`\n"+
			"│ • *Nominal:* Rp %s\n"+
			"│ • *Status:* ❌ TIMEOUT\n"+
			"└──────────────\n\n"+
			"*📝 Solusi:*\n"+
			"• Buat QRIS baru dengan `.donasi %d`\n"+
			"• Pastikan pembayaran dilakukan tepat waktu",
		refNo,
		helper.FormatAmount(amount),
		amount,
	)

	_, _ = ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
}


func getMustikaPayAPIKey() string {
	return mustikaPayAPIKey
}


var CekDonasiMetadata = &lib.CommandMetadata{
	Cmd:       "cekdonasi",
	Tag:       "main",
	Desc:      "Cek status donasi via Ref No",
	Example:   ".cekdonasi QR12345",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"cekdon", "statusdonasi"},
}


func CekDonasiHandler(ctx *lib.CommandContext) error {

	if len(ctx.Args) == 0 {

		dbManager, ok := ctx.BotClient.GetDBManager().(*helper.DatabaseManager)
		if ok && dbManager != nil {
			donations, err := dbManager.GetUserDonations(ctx.Sender.String())
			if err == nil && len(donations) > 0 {

				return showUserDonationHistory(ctx, donations)
			}
		}

		message := "*🔍 Cek Status Donasi*\n\n" +
			"┌─⦿ *Usage*\n" +
			"│ • `.cekdonasi <ref_no>` - Cek status via Ref No\n" +
			"│ • `.cekdonasi` - Lihat riwayat donasi Anda\n" +
			"└──────────────\n\n" +
			"*📝 Contoh:*\n" +
			"• `.cekdonasi QR12345`\n" +
			"• `.cekdonasi` (tanpa argumen)"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	refNo := ctx.Args[0]


	dbManager, ok := ctx.BotClient.GetDBManager().(*helper.DatabaseManager)
	if !ok || dbManager == nil {
		return fmt.Errorf("database manager tidak tersedia")
	}


	donation, err := dbManager.GetDonation(refNo)
	if err != nil {
		message := "❌ *Donasi tidak ditemukan!*\n\n" +
			"┌─⦿ *Info*\n" +
			fmt.Sprintf("│ • Ref No: %s\n", refNo) +
			"│ • Pastikan Ref No benar\n" +
			"└──────────────\n\n" +
			"*📝 Solusi:*\n" +
			"• Cek kembali Ref No Anda\n" +
			"• Gunakan `.cekdonasi` untuk lihat riwayat"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	if donation.UserJID != ctx.Sender.String() {
		message := "❌ *Akses Ditolak!*\n\n" +
			"┌─⦿ *Info*\n" +
			"│ • Donasi ini bukan milik Anda\n" +
			"└──────────────\n\n" +
			"*📝 Solusi:*\n" +
			"• Gunakan `.cekdonasi` untuk lihat donasi Anda"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	apiKey := getMustikaPayAPIKey()
	if apiKey == "" {
		return showDonationFromDatabase(ctx, donation)
	}


	logger := helper.NewLogger("CekDonasi")
	mustikaClient := helper.NewMustikaPayClient(apiKey, logger)


	status, err := mustikaClient.CheckPaymentStatus(refNo)
	if err != nil {
		return showDonationFromDatabase(ctx, donation)
	}


	if status.Status != donation.Status {
		dbManager.UpdateDonationStatus(refNo, helper.DonationStatus(status.Status), status.Type, status.Issuer, status.Payor)
		donation.Status = status.Status
		donation.PaymentType = status.Type
		donation.Issuer = status.Issuer
		donation.Payor = status.Payor
	}

	return showDonationFromDatabase(ctx, donation)
}


func showDonationFromDatabase(ctx *lib.CommandContext, donation *helper.DonationInfo) error {
	statusEmoji := "⏳"
	statusText := "PENDING"

	switch helper.DonationStatus(donation.Status) {
	case helper.DonationSuccess:
		statusEmoji = "✅"
		statusText = "SUCCESS"
	case helper.DonationExpired:
		statusEmoji = "⏰"
		statusText = "EXPIRED"
	case helper.DonationFailed:
		statusEmoji = "❌"
		statusText = "FAILED"
	}

	message := fmt.Sprintf(
		"%s *Status Donasi*\n\n"+
			"┌─⦿ *Detail*\n"+
			"│ • *Ref No:* `%s`\n"+
			"│ • *Nominal:* Rp %s\n"+
			"│ • *Produk:* %s\n"+
			"│ • *Status:* %s %s\n",
		statusEmoji,
		donation.RefNo,
		helper.FormatAmount(donation.Amount),
		donation.ProductName,
		statusEmoji,
		statusText,
	)

	if donation.PaymentType != "" {
		message += fmt.Sprintf("│ • *Metode:* %s\n", donation.PaymentType)
	}
	if donation.Issuer != "" {
		message += fmt.Sprintf("│ • *Issuer:* %s\n", donation.Issuer)
	}
	if donation.Payor != "" {
		message += fmt.Sprintf("│ • *Pembayar:* %s\n", donation.Payor)
	}

	message += "└──────────────\n"

	if helper.DonationStatus(donation.Status) == helper.DonationPending {
		message += "\n💡 *Tips:*\n" +
			"• Bot akan otomatis mendeteksi pembayaran\n" +
			"• QRIS berlaku selama 15 menit\n" +
			"• Cek kembali dalam beberapa saat"
	}

	_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return err
}


func showUserDonationHistory(ctx *lib.CommandContext, donations []helper.DonationInfo) error {
	var totalSuccess int
	var totalAmount int
	var pendingCount int

	message := "*📋 Riwayat Donasi Anda*\n\n" +
		fmt.Sprintf("┌─⦿ *Total: %d donasi*\n\n", len(donations))

	for i, donation := range donations {
		if i >= 10 {
			message += fmt.Sprintf("│ • ... dan %d donasi lainnya\n", len(donations)-10)
			break
		}

		switch helper.DonationStatus(donation.Status) {
		case helper.DonationSuccess:
			totalSuccess++
			totalAmount += donation.Amount
			message += fmt.Sprintf("%d. ✅ Rp %s - %s\n", i+1, helper.FormatAmount(donation.Amount), donation.RefNo)
		case helper.DonationPending:
			pendingCount++
			message += fmt.Sprintf("%d. ⏳ Rp %s - %s\n", i+1, helper.FormatAmount(donation.Amount), donation.RefNo)
		default:
			message += fmt.Sprintf("%d. ❌ Rp %s - %s\n", i+1, helper.FormatAmount(donation.Amount), donation.RefNo)
		}
	}

	message += "└──────────────\n\n" +
		"┌─⦿ *Statistik*\n" +
		fmt.Sprintf("│ • ✅ Sukses: %d\n", totalSuccess) +
		fmt.Sprintf("│ • ⏳ Pending: %d\n", pendingCount) +
		fmt.Sprintf("│ • 💰 Total Donasi: Rp %s\n", helper.FormatAmount(totalAmount)) +
		"└──────────────\n\n" +
		"*📝 Command:*\n" +
		"• `.cekdonasi <ref_no>` - Cek detail donasi\n" +
		"• `.donasi <nominal>` - Buat donasi baru"

	_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return err
}


func SetMustikaPayAPIKey(apiKey string) {

	mustikaPayAPIKey = apiKey
}

var mustikaPayAPIKey string
