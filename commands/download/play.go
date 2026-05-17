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


var PlayMetadata = &lib.CommandMetadata{
	Cmd:       "play",
	Tag:       "play",
	Desc:      "Download dan kirim audio dari YouTube",
	Example:   ".play Multo Cup of Joe",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"ytmp3", "yta"},
}


type PlayChannel struct {
	Name string
}

func (c *PlayChannel) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		c.Name = s
		return nil
	}
	var obj struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	c.Name = obj.Name
	return nil
}

type PlayResponse struct {
	Success bool   `json:"success"`
	Author  string `json:"author"`
	Result  struct {
		Title     string      `json:"title"`
		Thumbnail string      `json:"thumbnail"`
		URL       string      `json:"url_original"`
		SourceURL string      `json:"source_url"`
		Channel   PlayChannel `json:"channel"`
		Downloads []struct {
			Type     string `json:"type"`
			Quality  string `json:"quality"`
			Ext      string `json:"ext"`
			Duration string `json:"duration"`
			URL      string `json:"url"`
		} `json:"downloads"`
		Media struct {
			MediaExtension string `json:"mediaExtension"`
			MediaUrl       string `json:"mediaUrl"`
			Quality        string `json:"quality"`
		} `json:"media"`
	} `json:"result"`
}


func PlayHandler(ctx *lib.CommandContext) error {

	var query string
	if len(ctx.Args) > 0 {
		query = joinStrings(ctx.Args, " ")
	} else if ctx.ReplyMessage != nil {
		query = helper.ExtractMatchingURL(ctx.ReplyMessage.Message, helper.IsYouTubeURL)
	}

	if query == "" {
		message := "Masukkan judul atau link YouTube.\n\n" +
			"Usage:\n" +
			"- .play <judul> - Cari dan download audio\n" +
			"- .play <url> - Download dari URL YouTube\n" +
			"- Atau reply pesan berisi link YouTube dengan .play\n\n" +
			"Contoh:\n" +
			"- .play Multo Cup of Joe\n" +
			"- .play https://youtube.com/watch?v=xxxxx"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	if helper.LooksLikeURL(query) && !helper.IsYouTubeURL(query) {
		message := "URL YouTube tidak valid.\n\n" +
			"- Pakai link youtube.com / youtu.be, atau cari pakai keyword\n" +
			"- Contoh: .play https://youtu.be/xxxxx\n" +
			"- Contoh: .play lagu galau"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	apiURL := "https://api.yardansh.com/downloader/youtube-play?q=" + url.QueryEscape(query)

	playResp, err := fetchPlayAPI(apiURL)
	if err != nil {
		errorMsg := "Gagal mengambil data.\n\n" +
			fmt.Sprintf("Error: %s", err.Error())
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}


	if !playResp.Success || playResp.Result.Media.MediaUrl == "" {
		errorMsg := "Audio tidak ditemukan.\n\n" +
			"- Coba dengan kata kunci lain\n" +
			"- Pastikan judul benar"
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}


	return sendPlayAudio(ctx, playResp)
}


func sendPlayAudio(ctx *lib.CommandContext, data *PlayResponse) error {
	result := data.Result

	audioData, err := downloadFileFast(result.Media.MediaUrl)
	if err != nil {
		errorMsg := "Gagal download audio.\n\n" +
			fmt.Sprintf("Error: %s", err.Error())
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	uploadResp, err := ctx.Client.Upload(context.Background(), audioData, gowa.MediaAudio)
	if err != nil {
		return fmt.Errorf("failed to upload audio: %w", err)
	}

	var durationSeconds uint32
	for _, d := range result.Downloads {
		if d.Duration != "" {
			durationSeconds = parseDuration(d.Duration)
			break
		}
	}

	ptt := false

	audioMsg := &waE2E.Message{
		AudioMessage: &waE2E.AudioMessage{
			URL:               proto.String(uploadResp.URL),
			DirectPath:        proto.String(uploadResp.DirectPath),
			Mimetype:          proto.String("audio/mpeg"),
			PTT:               &ptt,
			FileSHA256:        uploadResp.FileSHA256,
			FileEncSHA256:     uploadResp.FileEncSHA256,
			FileLength:        proto.Uint64(uploadResp.FileLength),
			MediaKey:          uploadResp.MediaKey,
			MediaKeyTimestamp: proto.Int64(time.Now().Unix()),
			Seconds:           proto.Uint32(durationSeconds),
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:    proto.String(ctx.MessageID),
				Participant: proto.String(ctx.Sender.String()),
			},
		},
	}

	_, err = ctx.SendMessage(audioMsg)
	if err != nil {
		return fmt.Errorf("failed to send audio: %w", err)
	}

	return nil
}


func fetchPlayAPI(apiURL string) (*PlayResponse, error) {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}


	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://api.yardansh.com/")
	req.Header.Set("Origin", "https://api.yardansh.com")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch API: %w", err)
	}
	defer resp.Body.Close()


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


func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}


func parseDuration(duration string) uint32 {

	parts := splitString(duration, ":")
	if len(parts) == 3 {
		hours := parseUint(parts[0])
		minutes := parseUint(parts[1])
		seconds := parseUint(parts[2])
		return uint32(hours*3600 + minutes*60 + seconds)
	}
	if len(parts) == 2 {
		minutes := parseUint(parts[0])
		seconds := parseUint(parts[1])
		return uint32(minutes*60 + seconds)
	}

	return uint32(parseUint(duration))
}


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


func parseUint(s string) uint32 {
	var result uint32
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + uint32(c-'0')
		}
	}
	return result
}


func downloadFileFast(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 120 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "audio/*,*/*;q=0.9")

	startTime := time.Now()
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

	elapsed := time.Since(startTime)
	_ = elapsed

	return data, nil
}


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


func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
