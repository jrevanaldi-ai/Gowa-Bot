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


var SpotifyMetadata = &lib.CommandMetadata{
	Cmd:       "spotify",
	Tag:       "play",
	Desc:      "Download audio dari Spotify",
	Example:   ".sp Multo Cup of Joe",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"sp", "splay"},
}


type SpotifyResponse struct {
	Success bool         `json:"success"`
	Status  bool         `json:"status"`
	Author  string       `json:"author"`
	Result  SpotifyTrack `json:"result"`
}


type SpotifyTrack struct {
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	Album       string `json:"album"`
	ReleaseDate string `json:"release_date"`
	Duration    string `json:"duration"`
	Thumbnail   string `json:"thumbnail"`
	SpotifyUrl  string `json:"spotify_url"`
	DownloadUrl string `json:"download_url"`
}


func SpotifyHandler(ctx *lib.CommandContext) error {

	var query string
	if len(ctx.Args) > 0 {
		query = joinStrings(ctx.Args, " ")
	} else if ctx.ReplyMessage != nil {
		query = helper.ExtractMatchingURL(ctx.ReplyMessage.Message, helper.IsSpotifyURL)
	}

	if query == "" {
		message := "Masukkan judul lagu.\n\n" +
			"Usage:\n" +
			"- .sp <judul> - Cari dan download dari Spotify\n" +
			"- Atau reply pesan berisi link Spotify dengan .sp\n\n" +
			"Contoh:\n" +
			"- .sp Multo Cup of Joe\n" +
			"- .sp https://open.spotify.com/track/xxxx"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	if helper.LooksLikeURL(query) && !helper.IsSpotifyURL(query) {
		message := "URL Spotify tidak valid.\n\n" +
			"- Pakai link open.spotify.com/track/...\n" +
			"- Atau cari pakai keyword\n" +
			"- Contoh: .spotify https://open.spotify.com/track/xxxx"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	apiURL := "https://api.yardansh.com/downloader/spotify-play?q=" + url.QueryEscape(query)

	spResp, err := fetchSpotifyAPI(apiURL)
	if err != nil {
		errorMsg := "Gagal mengambil data.\n\n" +
			fmt.Sprintf("Error: %s", err.Error())
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}


	if !spResp.Success || spResp.Result.DownloadUrl == "" {
		errorMsg := "Lagu tidak ditemukan.\n\n" +
			"- Coba dengan kata kunci lain\n" +
			"- Pastikan judul benar"
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}


	return sendSpotifyAudio(ctx, spResp)
}


func fetchSpotifyAPI(apiURL string) (*SpotifyResponse, error) {
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
		return nil, fmt.Errorf("failed to fetch body: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
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


func sendSpotifyAudio(ctx *lib.CommandContext, data *SpotifyResponse) error {
	result := data.Result

	sendMsg := fmt.Sprintf(
		"Downloading audio...\n\n"+
			"Title: %s\n"+
			"Artist: %s",
		result.Title,
		result.Artist,
	)
	_, _ = ctx.SendMessage(helper.CreateSimpleReply(sendMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))


	audioData, err := downloadFileFast(result.DownloadUrl)
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


	durationSeconds := parseDuration(result.Duration)


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
