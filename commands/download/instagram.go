package download

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/jrevanaldi-ai/gowa"
	"github.com/jrevanaldi-ai/gowa/proto/waE2E"
	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

// InstagramMetadata adalah metadata untuk command instagram
var InstagramMetadata = &lib.CommandMetadata{
	Cmd:       "instagram",
	Tag:       "download",
	Desc:      "Download video atau foto dari Instagram",
	Example:   ".ig https://www.instagram.com/reel/xxxxx",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"ig", "igdl", "reels"},
}

// InstagramResponse struktur response dari API Instagram
type InstagramResponse struct {
	Creator string   `json:"creator"`
	Source  string   `json:"source"`
	Status  bool     `json:"status"`
	Type    string   `json:"type"`
	Message string   `json:"message"`
	Thumb   string   `json:"thumb"`
	Videos  []string `json:"videos"`
	Images  []string `json:"images"`
}

// InstagramHandler menangani command instagram
func InstagramHandler(ctx *lib.CommandContext) error {
	// Cek apakah ada argument
	if len(ctx.Args) == 0 {
		message := "❌ *Masukkan link Instagram!*\n\n" +
			"┌─⦿ *Usage*\n" +
			"│ • `.ig <url>` - Download dari Instagram\n" +
			"└──────────────\n\n" +
			"*📝 Contoh:*\n" +
			"• `.ig https://www.instagram.com/reel/xxxxx`\n" +
			"• `.ig https://www.instagram.com/p/xxxxx`"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Join semua args menjadi URL
	igURL := joinStrings(ctx.Args, " ")

	// Validasi URL Instagram
	if !isInstagramURL(igURL) {
		message := "❌ *URL Instagram tidak valid!*\n\n" +
			"┌─⦿ *Info*\n" +
			"│ • Pastikan URL dari instagram.com\n" +
			"│ • Contoh: `.ig https://www.instagram.com/reel/xxxxx`\n" +
			"└──────────────"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Fetch dari API
	apiURL := "https://api.azbry.com/api/download/instagram?url=" + url.QueryEscape(igURL)

	igResp, err := fetchInstagramAPI(apiURL)
	if err != nil {
		errorMsg := "❌ *Gagal mengambil data!*\n\n" +
			"┌─⦿ *Error*\n" +
			fmt.Sprintf("│ • %s\n", err.Error()) +
			"└──────────────"
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	// Validasi response
	if !igResp.Status {
		errorMsg := "❌ *Gagal download!*\n\n" +
			"┌─⦿ *Info*\n" +
			fmt.Sprintf("│ • %s\n", igResp.Message) +
			"└──────────────"
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	// Kirim media berdasarkan type
	return sendInstagramMedia(ctx, igResp)
}

// fetchInstagramAPI mengambil data dari API Instagram
func fetchInstagramAPI(apiURL string) (*InstagramResponse, error) {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var igResp InstagramResponse
	if err := json.Unmarshal(body, &igResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &igResp, nil
}

// sendInstagramMedia mengirim media dari Instagram
func sendInstagramMedia(ctx *lib.CommandContext, data *InstagramResponse) error {
	if data.Type == "video" && len(data.Videos) > 0 {
		return sendInstagramVideo(ctx, data)
	} else if data.Type == "image" && len(data.Images) > 0 {
		return sendInstagramImages(ctx, data)
	}

	errorMsg := "❌ *Media tidak ditemukan!*\n\n" +
		"┌─⦿ *Info*\n" +
		"│ • Tidak ada video atau foto yang bisa didownload\n" +
		"└──────────────"
	_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return nil
}

// sendInstagramVideo mengirim video dari Instagram
func sendInstagramVideo(ctx *lib.CommandContext, data *InstagramResponse) error {
	videoURL := data.Videos[0]

	// Download video
	videoData, err := downloadFileFast(videoURL)
	if err != nil {
		errorMsg := "❌ *Gagal download video!*\n\n" +
			"┌─⦿ *Error*\n" +
			fmt.Sprintf("│ • %s\n", err.Error()) +
			"└──────────────"
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	// Upload ke WhatsApp
	uploadResp, err := ctx.Client.Upload(context.Background(), videoData, gowa.MediaVideo)
	if err != nil {
		return fmt.Errorf("failed to upload video: %w", err)
	}

	// Buat pesan video
	senderStr := ctx.Sender.String()
	mediaType := waE2E.ContextInfo_ExternalAdReplyInfo_IMAGE
	adType := waE2E.ContextInfo_ExternalAdReplyInfo_CTWA
	showAd := true
	renderLarge := true

	videoMsg := &waE2E.Message{
		VideoMessage: &waE2E.VideoMessage{
			URL:           proto.String(uploadResp.URL),
			DirectPath:    proto.String(uploadResp.DirectPath),
			Mimetype:      proto.String("video/mp4"),
			Caption:       proto.String("📹 Instagram Reel"),
			FileSHA256:    uploadResp.FileSHA256,
			FileEncSHA256: uploadResp.FileEncSHA256,
			FileLength:    proto.Uint64(uploadResp.FileLength),
			MediaKey:      uploadResp.MediaKey,
			MediaKeyTimestamp: proto.Int64(time.Now().Unix()),
			Seconds:       proto.Uint32(0),
			GifPlayback:   proto.Bool(false),
			ContextInfo: &waE2E.ContextInfo{
				ExternalAdReply: &waE2E.ContextInfo_ExternalAdReplyInfo{
					Title:                 proto.String("Instagram Reel"),
					Body:                  proto.String("Video dari Instagram"),
					MediaType:             &mediaType,
					ThumbnailURL:          &data.Thumb,
					ShowAdAttribution:     &showAd,
					RenderLargerThumbnail: &renderLarge,
					AdType:                &adType,
				},
				StanzaID:    &ctx.MessageID,
				Participant: &senderStr,
			},
		},
	}

	_, err = ctx.SendMessage(videoMsg)
	if err != nil {
		return fmt.Errorf("failed to send video: %w", err)
	}

	return nil
}

// sendInstagramImages mengirim gambar dari Instagram (bisa multiple)
func sendInstagramImages(ctx *lib.CommandContext, data *InstagramResponse) error {
	for i, imgURL := range data.Images {
		// Download gambar
		imgData, err := downloadFileFast(imgURL)
		if err != nil {
			continue
		}

		// Upload ke WhatsApp
		uploadResp, err := ctx.Client.Upload(context.Background(), imgData, gowa.MediaImage)
		if err != nil {
			continue
		}

		// Buat pesan gambar
		senderStr := ctx.Sender.String()
		mediaType := waE2E.ContextInfo_ExternalAdReplyInfo_IMAGE
		adType := waE2E.ContextInfo_ExternalAdReplyInfo_CTWA
		showAd := true
		renderLarge := true

		caption := fmt.Sprintf("📸 Instagram Photo (%d/%d)", i+1, len(data.Images))

		imageMsg := &waE2E.Message{
			ImageMessage: &waE2E.ImageMessage{
				URL:           proto.String(uploadResp.URL),
				DirectPath:    proto.String(uploadResp.DirectPath),
				Mimetype:      proto.String("image/jpeg"),
				Caption:       proto.String(caption),
				FileSHA256:    uploadResp.FileSHA256,
				FileEncSHA256: uploadResp.FileEncSHA256,
				FileLength:    proto.Uint64(uploadResp.FileLength),
				MediaKey:      uploadResp.MediaKey,
				MediaKeyTimestamp: proto.Int64(time.Now().Unix()),
				ContextInfo: &waE2E.ContextInfo{
					ExternalAdReply: &waE2E.ContextInfo_ExternalAdReplyInfo{
						Title:                 proto.String("Instagram Photo"),
						Body:                  proto.String("Foto dari Instagram"),
						MediaType:             &mediaType,
						ShowAdAttribution:     &showAd,
						RenderLargerThumbnail: &renderLarge,
						AdType:                &adType,
					},
					StanzaID:    &ctx.MessageID,
					Participant: &senderStr,
				},
			},
		}

		_, err = ctx.SendMessage(imageMsg)
		if err != nil {
			return fmt.Errorf("failed to send image %d: %w", i+1, err)
		}
	}

	return nil
}

// isInstagramURL cek apakah URL valid dari Instagram
func isInstagramURL(url string) bool {
	return containsString(url, "instagram.com/")
}

// containsString cek apakah string mengandung substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && s != "" && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

// containsSubstring helper
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
