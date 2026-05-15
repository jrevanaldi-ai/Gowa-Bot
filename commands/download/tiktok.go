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


var TikTokMetadata = &lib.CommandMetadata{
	Cmd:       "tiktok",
	Tag:       "download",
	Desc:      "Download video dari TikTok",
	Example:   ".tt https://vt.tiktok.com/xxxxx",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"tt", "tikdl", "tiktokdl"},
}


type TikTokResponse struct {
	Status  bool       `json:"status"`
	Result  TikTokData `json:"result"`
	Message string     `json:"message"`
}


type TikTokData struct {
	Title     string   `json:"title"`
	Author    string   `json:"author"`
	Thumbnail string   `json:"thumbnail"`
	Links     []string `json:"links"`
}


func TikTokHandler(ctx *lib.CommandContext) error {

	var ttURL string
	if len(ctx.Args) > 0 {
		ttURL = strings.Join(ctx.Args, " ")
	} else if ctx.ReplyMessage != nil {
		ttURL = helper.ExtractMatchingURL(ctx.ReplyMessage.Message, helper.IsTikTokURL)
	}

	if ttURL == "" {
		message := "Masukkan link TikTok dulu.\n\n" +
			"Usage:\n" +
			"- .tt <url> - Download dari TikTok\n" +
			"- Atau reply pesan berisi link TikTok dengan .tt\n\n" +
			"Contoh:\n" +
			"- .tt https://vt.tiktok.com/xxxxx\n" +
			"- .tt https://www.tiktok.com/@user/video/xxxxx"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	if !helper.IsTikTokURL(ttURL) {
		message := "URL TikTok tidak valid.\n\n" +
			"- Pastikan URL dari tiktok.com / vt.tiktok.com\n" +
			"- Contoh: .tt https://vt.tiktok.com/xxxxx"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	apiURL := "https://api.azbry.com/api/download/tiktok?url=" + url.QueryEscape(ttURL)

	ttResp, err := fetchTikTokAPI(apiURL)
	if err != nil {
		errorMsg := "Gagal mengambil data.\n\n" +
			fmt.Sprintf("Error: %s", err.Error())
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}


	if !ttResp.Status || len(ttResp.Result.Links) == 0 {
		errorMsg := "Gagal download.\n\n" +
			fmt.Sprintf("- %s\n", ttResp.Message) +
			"- Pastikan link TikTok valid"
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}


	return sendTikTokVideo(ctx, ttResp)
}


func fetchTikTokAPI(apiURL string) (*TikTokResponse, error) {
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

	var ttResp TikTokResponse
	if err := json.Unmarshal(body, &ttResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &ttResp, nil
}


func sendTikTokVideo(ctx *lib.CommandContext, data *TikTokResponse) error {

	videoURL := data.Result.Links[0]


	if !strings.HasPrefix(videoURL, "http://") && !strings.HasPrefix(videoURL, "https://") {
		videoURL = "https://" + videoURL
	}


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
	mediaType := waE2E.ContextInfo_ExternalAdReplyInfo_IMAGE
	adType := waE2E.ContextInfo_ExternalAdReplyInfo_CTWA
	showAd := true
	renderLarge := true

	title := data.Result.Title
	if title == "" {
		title = "TikTok Video"
	}

	author := data.Result.Author
	if author == "" {
		author = "Unknown"
	}

	caption := fmt.Sprintf("%s\n%s", title, author)

	videoMsg := &waE2E.Message{
		VideoMessage: &waE2E.VideoMessage{
			URL:               proto.String(uploadResp.URL),
			DirectPath:        proto.String(uploadResp.DirectPath),
			Mimetype:          proto.String("video/mp4"),
			Caption:           proto.String(caption),
			FileSHA256:        uploadResp.FileSHA256,
			FileEncSHA256:     uploadResp.FileEncSHA256,
			FileLength:        proto.Uint64(uploadResp.FileLength),
			MediaKey:          uploadResp.MediaKey,
			MediaKeyTimestamp: proto.Int64(time.Now().Unix()),
			Seconds:           proto.Uint32(0),
			GifPlayback:       proto.Bool(false),
			ContextInfo: &waE2E.ContextInfo{
				ExternalAdReply: &waE2E.ContextInfo_ExternalAdReplyInfo{
					Title:                 proto.String("TikTok Video"),
					Body:                  proto.String(author),
					MediaType:             &mediaType,
					ThumbnailURL:          &data.Result.Thumbnail,
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
