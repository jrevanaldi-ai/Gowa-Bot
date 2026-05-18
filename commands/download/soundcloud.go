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

var SoundCloudMetadata = &lib.CommandMetadata{
	Cmd:       "soundcloud",
	Tag:       "download",
	Desc:      "Download audio dari SoundCloud",
	Example:   ".sc https://soundcloud.com/user/track",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"sc", "scdl"},
}

type SoundCloudResponse struct {
	Status bool             `json:"status"`
	Data   SoundCloudTrack  `json:"data"`
}

type SoundCloudTrack struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Thumbnail   string `json:"thumbnail"`
	Duration    uint64 `json:"duration"`
	User        string `json:"user"`
	Description string `json:"description"`
}

func SoundCloudHandler(ctx *lib.CommandContext) error {
	var scURL string
	if len(ctx.Args) > 0 {
		scURL = joinStrings(ctx.Args, " ")
	} else if ctx.ReplyMessage != nil {
		scURL = helper.ExtractMatchingURL(ctx.ReplyMessage.Message, helper.IsSoundCloudURL)
	}

	if scURL == "" {
		message := "Masukkan link SoundCloud.\n\n" +
			"Usage:\n" +
			"- .sc <url> - Download audio dari SoundCloud\n" +
			"- Atau reply pesan berisi link SoundCloud dengan .sc\n\n" +
			"Contoh:\n" +
			"- .sc https://soundcloud.com/user/track\n" +
			"- .sc https://m.soundcloud.com/user/track"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	if !helper.IsSoundCloudURL(scURL) {
		message := "URL SoundCloud tidak valid.\n\n" +
			"- Format: https://soundcloud.com/<user>/<track>\n" +
			"- Mobile: https://m.soundcloud.com/<user>/<track>\n" +
			"- Short: https://on.soundcloud.com/<id>"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	apiURL := "https://api.siputzx.my.id/api/d/soundcloud?url=" + url.QueryEscape(scURL)

	scResp, err := fetchSoundCloudAPI(apiURL)
	if err != nil {
		errorMsg := "Gagal mengambil data.\n\n" +
			fmt.Sprintf("Error: %s", err.Error())
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	if !scResp.Status || scResp.Data.URL == "" {
		errorMsg := "Track tidak ditemukan.\n\n" +
			"- Pastikan URL valid dan track public\n" +
			"- Track private/locked tidak bisa di-download"
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	return sendSoundCloudAudio(ctx, scResp)
}

func fetchSoundCloudAPI(apiURL string) (*SoundCloudResponse, error) {
	client := &http.Client{
		Timeout: 90 * time.Second,
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var scResp SoundCloudResponse
	if err := json.Unmarshal(body, &scResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &scResp, nil
}

func sendSoundCloudAudio(ctx *lib.CommandContext, data *SoundCloudResponse) error {
	track := data.Data

	durationSec := uint32(track.Duration / 1000)
	durationDisplay := formatDurationFromSec(durationSec)

	loadingMsg := fmt.Sprintf(
		"Downloading audio...\n\n"+
			"Title: %s\n"+
			"User: %s\n"+
			"Duration: %s",
		track.Title,
		track.User,
		durationDisplay,
	)
	_, _ = ctx.SendMessage(helper.CreateSimpleReply(loadingMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))

	audioData, err := downloadFileFast(track.URL)
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
			Seconds:           proto.Uint32(durationSec),
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:    proto.String(ctx.MessageID),
				Participant: proto.String(ctx.Sender.String()),
			},
		},
	}

	if _, err := ctx.SendMessage(audioMsg); err != nil {
		return fmt.Errorf("failed to send audio: %w", err)
	}
	return nil
}

func formatDurationFromSec(sec uint32) string {
	if sec == 0 {
		return "-"
	}
	h := sec / 3600
	m := (sec % 3600) / 60
	s := sec % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}
