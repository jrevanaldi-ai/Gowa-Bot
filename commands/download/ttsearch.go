package download

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/jrevanaldi-ai/gowa"
	"github.com/jrevanaldi-ai/gowa/proto/waE2E"
	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

// TTSearchMetadata adalah metadata untuk command tiktok search
var TTSearchMetadata = &lib.CommandMetadata{
	Cmd:       "ttsearch",
	Tag:       "search",
	Desc:      "Cari dan download video dari TikTok",
	Example:   ".tts ayam",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"tts", "tiktoksearch"},
}

// TTSearchResponse struktur response dari API TikTok Search
type TTSearchResponse struct {
	Creator string         `json:"creator"`
	Status  bool           `json:"status"`
	Result  []TTSearchItem `json:"result"`
}

// TTSearchItem item video dari TikTok search
type TTSearchItem struct {
	Title    string          `json:"title"`
	Cover    string          `json:"cover"`
	Link     string          `json:"link"`
	Author   TTSearchAuthor  `json:"author"`
	Stats    TTSearchStats   `json:"stats"`
}

// TTSearchAuthor info author TikTok
type TTSearchAuthor struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// TTSearchStats statistik video TikTok
type TTSearchStats struct {
	Plays    int `json:"plays"`
	Likes    int `json:"likes"`
	Comments int `json:"comments"`
	Shares   int `json:"shares"`
}

// TTSearchHandler menangani command ttsearch
func TTSearchHandler(ctx *lib.CommandContext) error {
	// Cek apakah ada argument
	if len(ctx.Args) == 0 {
		message := "❌ *Masukkan kata kunci pencarian!*\n\n" +
			"┌─⦿ *Usage*\n" +
			"│ • `.tts <keyword>` - Cari video TikTok\n" +
			"└──────────────\n\n" +
			"*📝 Contoh:*\n" +
			"• `.tts ayam`\n" +
			"• `.tts resep mudah`\n" +
			"• `.tts kucing lucu`"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Join semua args menjadi query
	query := joinStrings(ctx.Args, " ")

	// Fetch dari API
	apiURL := "https://api.azbry.com/api/search/ttsearch?q=" + url.QueryEscape(query)

	searchResp, err := fetchTTSearchAPI(apiURL)
	if err != nil {
		errorMsg := "❌ *Gagal mengambil data!*\n\n" +
			"┌─⦿ *Error*\n" +
			fmt.Sprintf("│ • %s\n", err.Error()) +
			"└──────────────"
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	// Validasi response
	if !searchResp.Status || len(searchResp.Result) == 0 {
		errorMsg := "❌ *Video tidak ditemukan!*\n\n" +
			"┌─⦿ *Info*\n" +
			"│ • Coba dengan kata kunci lain\n" +
			"└──────────────"
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	// Ambil video pertama dan kirim
	return sendTTSearchVideo(ctx, searchResp.Result[0], 1, len(searchResp.Result))
}

// fetchTTSearchAPI mengambil data dari API TikTok Search
func fetchTTSearchAPI(apiURL string) (*TTSearchResponse, error) {
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

	var searchResp TTSearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &searchResp, nil
}

// sendTTSearchVideo mengirim video dari hasil search
func sendTTSearchVideo(ctx *lib.CommandContext, item TTSearchItem, index int, total int) error {
	// Bersihkan link - kadang ada prefix "https://tikwm.com" atau "http://tikwm.com"
	videoURL := item.Link
	videoURL = strings.TrimPrefix(videoURL, "https://tikwm.com")
	videoURL = strings.TrimPrefix(videoURL, "http://tikwm.com")
	videoURL = strings.TrimSpace(videoURL)

	// Pastikan URL valid (harus dimulai dengan http:// atau https://)
	if !strings.HasPrefix(videoURL, "http://") && !strings.HasPrefix(videoURL, "https://") {
		videoURL = "https://" + videoURL
	}

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

	// Buat caption
	title := truncateString(item.Title, 50)
	author := item.Author.Nickname
	views := formatNumber(item.Stats.Plays)
	likes := formatNumber(item.Stats.Likes)

	caption := fmt.Sprintf("🎵 %s\n\n👤 %s\n👁️ %s views | ❤️ %s likes", title, author, views, likes)

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
			Caption:       proto.String(caption),
			FileSHA256:    uploadResp.FileSHA256,
			FileEncSHA256: uploadResp.FileEncSHA256,
			FileLength:    proto.Uint64(uploadResp.FileLength),
			MediaKey:      uploadResp.MediaKey,
			MediaKeyTimestamp: proto.Int64(time.Now().Unix()),
			Seconds:       proto.Uint32(0),
			GifPlayback:   proto.Bool(false),
			ContextInfo: &waE2E.ContextInfo{
				ExternalAdReply: &waE2E.ContextInfo_ExternalAdReplyInfo{
					Title:                 proto.String("TikTok Search"),
					Body:                  proto.String(fmt.Sprintf("%s - %s", author, views)),
					MediaType:             &mediaType,
					ThumbnailURL:          &item.Cover,
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

// formatNumber memformat angka menjadi lebih readable
func formatNumber(num int) string {
	if num >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(num)/1000000)
	} else if num >= 1000 {
		return fmt.Sprintf("%.1fK", float64(num)/1000)
	}
	return fmt.Sprintf("%d", num)
}
