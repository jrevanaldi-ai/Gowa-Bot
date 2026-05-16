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


var InstagramMetadata = &lib.CommandMetadata{
	Cmd:       "instagram",
	Tag:       "download",
	Desc:      "Download video atau foto dari Instagram",
	Example:   ".ig https://www.instagram.com/reel/xxxxx",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"ig", "igdl", "reels"},
}


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


func InstagramHandler(ctx *lib.CommandContext) error {

	var igURL string
	if len(ctx.Args) > 0 {
		igURL = joinStrings(ctx.Args, " ")
	} else if ctx.ReplyMessage != nil {
		igURL = helper.ExtractMatchingURL(ctx.ReplyMessage.Message, helper.IsInstagramURL)
	}

	if igURL == "" {
		message := "Masukkan link Instagram.\n\n" +
			"Usage:\n" +
			"- .ig <url> - Download dari Instagram\n" +
			"- Atau reply pesan berisi link Instagram dengan .ig\n\n" +
			"Contoh:\n" +
			"- .ig https://www.instagram.com/reel/xxxxx\n" +
			"- .ig https://www.instagram.com/p/xxxxx"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	if !helper.IsInstagramURL(igURL) {
		message := "URL Instagram tidak valid.\n\n" +
			"- Format: https://www.instagram.com/p/xxxx atau /reel/xxxx\n" +
			"- Contoh: .ig https://www.instagram.com/reel/xxxxx"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	apiURL := "https://api.azbry.com/api/download/instagram?url=" + url.QueryEscape(igURL)

	igResp, err := fetchInstagramAPI(apiURL)
	if err != nil {
		errorMsg := "Gagal mengambil data.\n\n" +
			fmt.Sprintf("Error: %s", err.Error())
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}


	if !igResp.Status {
		errorMsg := "Gagal download.\n\n" +
			fmt.Sprintf("- %s", igResp.Message)
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}


	return sendInstagramMedia(ctx, igResp)
}


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


func sendInstagramMedia(ctx *lib.CommandContext, data *InstagramResponse) error {
	if data.Type == "video" && len(data.Videos) > 0 {
		return sendInstagramVideo(ctx, data)
	} else if data.Type == "image" && len(data.Images) > 0 {
		return sendInstagramImages(ctx, data)
	}

	errorMsg := "Media tidak ditemukan.\n\n" +
		"- Tidak ada video atau foto yang bisa didownload"
	_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return nil
}


func sendInstagramVideo(ctx *lib.CommandContext, data *InstagramResponse) error {
	videoURL := data.Videos[0]


	videoData, err := downloadFileFast(videoURL)
	if err != nil {
		errorMsg := "Gagal download video.\n\n" +
			fmt.Sprintf("Error: %s", err.Error())
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}


	uploadResp, err := ctx.Client.Upload(context.Background(), videoData, gowa.MediaVideo)
	if err != nil {
		return fmt.Errorf("failed to upload video: %w", err)
	}


	senderStr := ctx.Sender.String()
	thumbBytes := helper.FetchThumbnail(data.Thumb)

	videoMsg := &waE2E.Message{
		VideoMessage: &waE2E.VideoMessage{
			URL:               proto.String(uploadResp.URL),
			DirectPath:        proto.String(uploadResp.DirectPath),
			Mimetype:          proto.String("video/mp4"),
			Caption:           proto.String("Instagram Reel"),
			FileSHA256:        uploadResp.FileSHA256,
			FileEncSHA256:     uploadResp.FileEncSHA256,
			FileLength:        proto.Uint64(uploadResp.FileLength),
			MediaKey:          uploadResp.MediaKey,
			MediaKeyTimestamp: proto.Int64(time.Now().Unix()),
			Seconds:           proto.Uint32(0),
			GifPlayback:       proto.Bool(false),
			JPEGThumbnail:     thumbBytes,
			ContextInfo: &waE2E.ContextInfo{
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


func sendInstagramImages(ctx *lib.CommandContext, data *InstagramResponse) error {
	for i, imgURL := range data.Images {

		imgData, err := downloadFileFast(imgURL)
		if err != nil {
			continue
		}


		uploadResp, err := ctx.Client.Upload(context.Background(), imgData, gowa.MediaImage)
		if err != nil {
			continue
		}


		senderStr := ctx.Sender.String()

		caption := fmt.Sprintf("Instagram Photo (%d/%d)", i+1, len(data.Images))

		imageMsg := &waE2E.Message{
			ImageMessage: &waE2E.ImageMessage{
				URL:               proto.String(uploadResp.URL),
				DirectPath:        proto.String(uploadResp.DirectPath),
				Mimetype:          proto.String("image/jpeg"),
				Caption:           proto.String(caption),
				FileSHA256:        uploadResp.FileSHA256,
				FileEncSHA256:     uploadResp.FileEncSHA256,
				FileLength:        proto.Uint64(uploadResp.FileLength),
				MediaKey:          uploadResp.MediaKey,
				MediaKeyTimestamp: proto.Int64(time.Now().Unix()),
				ContextInfo: &waE2E.ContextInfo{
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


