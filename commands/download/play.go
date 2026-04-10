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

// PlayMetadata adalah metadata untuk command play
var PlayMetadata = &lib.CommandMetadata{
	Cmd:       "play",
	Tag:       "download",
	Desc:      "Download dan kirim audio dari YouTube",
	Example:   ".play Multo Cup of Joe",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"ytmp3", "yta"},
}

// PlayResponse struktur response dari API
type PlayResponse struct {
	Creator string `json:"creator"`
	Source  string `json:"source"`
	Status  bool   `json:"status"`
	Result  struct {
		Title     string `json:"title"`
		Channel   string `json:"channel"`
		Thumbnail string `json:"thumbnail"`
		Duration  string `json:"duration"`
		VideoID   string `json:"videoId"`
		URL       string `json:"url"`
		Download  string `json:"download"`
		Format    string `json:"format"`
	} `json:"result"`
}

// PlayHandler menangani command play
func PlayHandler(ctx *lib.CommandContext) error {
	// Cek apakah ada argument
	if len(ctx.Args) == 0 {
		message := "❌ *Masukkan judul atau link YouTube!*\n\n" +
			"┌─⦿ *Usage*\n" +
			"│ • `.play <judul>` - Cari dan download audio\n" +
			"│ • `.play <url>` - Download dari URL YouTube\n" +
			"└──────────────\n\n" +
			"*📝 Contoh:*\n" +
			"• `.play Multo Cup of Joe`\n" +
			"• `.play https://youtube.com/watch?v=xxxxx`"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Join semua args menjadi query
	query := joinStrings(ctx.Args, " ")

	// Fetch dari API
	apiURL := "https://api.azbry.com/api/download/ytplay2?q=" + url.QueryEscape(query)

	playResp, err := fetchPlayAPI(apiURL)
	if err != nil {
		errorMsg := "❌ *Gagal mengambil data!*\n\n" +
			"┌─⦿ *Error*\n" +
			fmt.Sprintf("│ • %s\n", err.Error()) +
			"└──────────────"
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	// Validasi response
	if !playResp.Status {
		errorMsg := "❌ *Audio tidak ditemukan!*\n\n" +
			"┌─⦿ *Info*\n" +
			"│ • Coba dengan kata kunci lain\n" +
			"│ • Pastikan judul benar\n" +
			"└──────────────"
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	// Langsung download dan kirim audio
	return sendPlayAudio(ctx, playResp)
}

// sendPlayAudio mengirim audio dengan format ExternalAdReply
func sendPlayAudio(ctx *lib.CommandContext, data *PlayResponse) error {
	result := data.Result

	// Download audio dari API
	audioData, err := downloadFileFast(result.Download)
	if err != nil {
		errorMsg := "❌ *Gagal download audio!*\n\n" +
			"┌─⦿ *Error*\n" +
			fmt.Sprintf("│ • %s\n", err.Error()) +
			"└──────────────"
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	// Upload ke WhatsApp
	uploadResp, err := ctx.Client.Upload(context.Background(), audioData, gowa.MediaAudio)
	if err != nil {
		return fmt.Errorf("failed to upload audio: %w", err)
	}

	// Parse duration
	durationSeconds := parseDuration(result.Duration)

	// Buat pesan audio
	senderStr := ctx.Sender.String()
	mediaType := waE2E.ContextInfo_ExternalAdReplyInfo_IMAGE
	adType := waE2E.ContextInfo_ExternalAdReplyInfo_CTWA
	showAd := true
	renderLarge := true
	ptt := false

	audioMsg := &waE2E.Message{
		AudioMessage: &waE2E.AudioMessage{
			URL:           proto.String(uploadResp.URL),
			DirectPath:    proto.String(uploadResp.DirectPath),
			Mimetype:      proto.String("audio/mpeg"),
			PTT:           &ptt,
			FileSHA256:    uploadResp.FileSHA256,
			FileEncSHA256: uploadResp.FileEncSHA256,
			FileLength:    proto.Uint64(uploadResp.FileLength),
			MediaKey:      uploadResp.MediaKey,
			MediaKeyTimestamp: proto.Int64(time.Now().Unix()),
			Seconds:       proto.Uint32(durationSeconds),
			ContextInfo: &waE2E.ContextInfo{
				ExternalAdReply: &waE2E.ContextInfo_ExternalAdReplyInfo{
					Title:                 &result.Title,
					Body:                  proto.String(fmt.Sprintf("%s • %s", result.Channel, result.Duration)),
					MediaType:             &mediaType,
					ThumbnailURL:          &result.Thumbnail,
					SourceURL:             &result.URL,
					ShowAdAttribution:     &showAd,
					RenderLargerThumbnail: &renderLarge,
					AdType:                &adType,
				},
				StanzaID:    &ctx.MessageID,
				Participant: &senderStr,
			},
		},
	}

	_, err = ctx.SendMessage(audioMsg)
	if err != nil {
		return fmt.Errorf("failed to send audio: %w", err)
	}

	return nil
}

// fetchPlayAPI mengambil data dari API
func fetchPlayAPI(apiURL string) (*PlayResponse, error) {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers yang lebih lengkap
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://api.azbry.com/")
	req.Header.Set("Origin", "https://api.azbry.com")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch API: %w", err)
	}
	defer resp.Body.Close()

	// Read body untuk error message juga
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body[:min(len(body), 200)]))
	}

	var playResp PlayResponse
	if err := json.Unmarshal(body, &playResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &playResp, nil
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// parseDuration memparse durasi dari format "4:01" ke seconds
func parseDuration(duration string) uint32 {
	// Split by ":"
	parts := splitString(duration, ":")
	if len(parts) == 2 {
		minutes := parseUint(parts[0])
		seconds := parseUint(parts[1])
		return uint32(minutes*60 + seconds)
	}
	// Kalau hanya 1 angka (detik)
	return uint32(parseUint(duration))
}

// splitString sederhana split string
func splitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

// parseUint parse string ke uint
func parseUint(s string) uint32 {
	var result uint32
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + uint32(c-'0')
		}
	}
	return result
}

// downloadFileFast download file dengan timeout cepat
func downloadFileFast(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil // Allow all redirects
		},
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "audio/*,*/*;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	return data, nil
}

// joinStrings menggabungkan slice string dengan separator
func joinStrings(elems []string, sep string) string {
	if len(elems) == 0 {
		return ""
	}
	result := elems[0]
	for i := 1; i < len(elems); i++ {
		result += sep + elems[i]
	}
	return result
}

// truncateString memotong string jika terlalu panjang
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
