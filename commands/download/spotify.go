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

// SpotifyMetadata adalah metadata untuk command spotify
var SpotifyMetadata = &lib.CommandMetadata{
	Cmd:       "spotify",
	Tag:       "download",
	Desc:      "Download audio dari Spotify",
	Example:   ".sp Multo Cup of Joe",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"sp", "splay"},
}

// SpotifyResponse struktur response dari API Spotify
type SpotifyResponse struct {
	Creator string          `json:"creator"`
	Source  string          `json:"source"`
	Status  bool            `json:"status"`
	Query   string          `json:"query"`
	Result  SpotifyTrack    `json:"result"`
}

// SpotifyTrack informasi track dari Spotify
type SpotifyTrack struct {
	Title        string `json:"title"`
	Artist       string `json:"artist"`
	Album        string `json:"album"`
	Cover        string `json:"cover"`
	Duration     int    `json:"duration"`
	DeezerUrl    string `json:"deezerUrl"`
	DownloadLink string `json:"downloadLink"`
	RawLink      string `json:"rawLink"`
}

// SpotifyHandler menangani command spotify
func SpotifyHandler(ctx *lib.CommandContext) error {
	// Cek apakah ada argument
	if len(ctx.Args) == 0 {
		message := "❌ *Masukkan judul lagu!*\n\n" +
			"┌─⦿ *Usage*\n" +
			"│ • `.sp <judul>` - Cari dan download dari Spotify\n" +
			"└──────────────\n\n" +
			"*📝 Contoh:*\n" +
			"• `.sp Multo Cup of Joe`\n" +
			"• `.sp Wonderwall Oasis`"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Join semua args menjadi query
	query := joinStrings(ctx.Args, " ")

	// Fetch dari API
	apiURL := "https://api.azbry.com/api/download/spoplay?q=" + url.QueryEscape(query)

	spResp, err := fetchSpotifyAPI(apiURL)
	if err != nil {
		errorMsg := "❌ *Gagal mengambil data!*\n\n" +
			"┌─⦿ *Error*\n" +
			fmt.Sprintf("│ • %s\n", err.Error()) +
			"└──────────────"
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	// Validasi response
	if !spResp.Status || spResp.Result.DownloadLink == "" {
		errorMsg := "❌ *Lagu tidak ditemukan!*\n\n" +
			"┌─⦿ *Info*\n" +
			"│ • Coba dengan kata kunci lain\n" +
			"│ • Pastikan judul benar\n" +
			"└──────────────"
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	// Langsung download dan kirim audio
	return sendSpotifyAudio(ctx, spResp)
}

// fetchSpotifyAPI mengambil data dari API Spotify
func fetchSpotifyAPI(apiURL string) (*SpotifyResponse, error) {
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

	var spResp SpotifyResponse
	if err := json.Unmarshal(body, &spResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &spResp, nil
}

// sendSpotifyAudio mengirim audio dari Spotify
func sendSpotifyAudio(ctx *lib.CommandContext, data *SpotifyResponse) error {
	result := data.Result

	// Download audio dari API
	audioData, err := downloadFileFast(result.DownloadLink)
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

	// Parse duration (sudah dalam detik dari API)
	durationSeconds := uint32(result.Duration)

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
					Body:                  proto.String(fmt.Sprintf("%s • %s", result.Artist, result.Album)),
					MediaType:             &mediaType,
					ThumbnailURL:          &result.Cover,
					SourceURL:             &result.DeezerUrl,
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
