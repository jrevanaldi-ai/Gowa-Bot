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

var SnackVideoMetadata = &lib.CommandMetadata{
	Cmd:       "snackvideo",
	Tag:       "download",
	Desc:      "Download video dari SnackVideo",
	Example:   ".snack https://s.snackvideo.com/p/xxx",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"snack", "snackdl", "sv"},
}

type SnackVideoResponse struct {
	Status bool            `json:"status"`
	Data   SnackVideoTrack `json:"data"`
}

type SnackVideoTrack struct {
	URL         string                `json:"url"`
	Title       string                `json:"title"`
	Description string                `json:"description"`
	Thumbnail   string                `json:"thumbnail"`
	UploadDate  string                `json:"uploadDate"`
	VideoURL    string                `json:"videoUrl"`
	Duration    string                `json:"duration"`
	Interaction SnackVideoInteraction `json:"interaction"`
	Creator     SnackVideoCreator     `json:"creator"`
}

type SnackVideoInteraction struct {
	Views  uint64 `json:"views"`
	Likes  uint64 `json:"likes"`
	Shares uint64 `json:"shares"`
}

type SnackVideoCreator struct {
	Name       string `json:"name"`
	ProfileURL string `json:"profileUrl"`
	Bio        string `json:"bio"`
}

func SnackVideoHandler(ctx *lib.CommandContext) error {
	var svURL string
	if len(ctx.Args) > 0 {
		svURL = joinStrings(ctx.Args, " ")
	} else if ctx.ReplyMessage != nil {
		svURL = helper.ExtractMatchingURL(ctx.ReplyMessage.Message, helper.IsSnackVideoURL)
	}

	if svURL == "" {
		message := "Masukkan link SnackVideo.\n\n" +
			"Usage:\n" +
			"- .snack <url> - Download video dari SnackVideo\n" +
			"- Atau reply pesan berisi link dengan .snack\n\n" +
			"Contoh:\n" +
			"- .snack https://s.snackvideo.com/p/dwlMd51U\n" +
			"- .snack https://www.snackvideo.com/@user/video/123"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	if !helper.IsSnackVideoURL(svURL) {
		message := "URL SnackVideo tidak valid.\n\n" +
			"- Short: https://s.snackvideo.com/p/<id>\n" +
			"- Full:  https://www.snackvideo.com/@<user>/video/<id>"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	apiURL := "https://api.siputzx.my.id/api/d/snackvideo?url=" + url.QueryEscape(svURL)

	svResp, err := fetchSnackVideoAPI(apiURL)
	if err != nil {
		errorMsg := "Gagal mengambil data.\n\n" +
			fmt.Sprintf("Error: %s", err.Error())
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	if !svResp.Status || svResp.Data.VideoURL == "" {
		errorMsg := "Video tidak ditemukan.\n\n" +
			"- Pastikan URL valid\n" +
			"- Video private/terhapus tidak bisa di-download"
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	return sendSnackVideoMedia(ctx, svResp)
}

func fetchSnackVideoAPI(apiURL string) (*SnackVideoResponse, error) {
	client := &http.Client{Timeout: 60 * time.Second}

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

	var svResp SnackVideoResponse
	if err := json.Unmarshal(body, &svResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return &svResp, nil
}

func sendSnackVideoMedia(ctx *lib.CommandContext, data *SnackVideoResponse) error {
	track := data.Data

	caption := buildSnackVideoCaption(&track)

	videoData, err := downloadFileFast(track.VideoURL)
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
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:    proto.String(ctx.MessageID),
				Participant: proto.String(ctx.Sender.String()),
			},
		},
	}

	if _, err := ctx.SendMessage(videoMsg); err != nil {
		return fmt.Errorf("failed to send video: %w", err)
	}
	return nil
}

func buildSnackVideoCaption(t *SnackVideoTrack) string {
	desc := strings.TrimSpace(t.Description)
	if desc == "" {
		desc = strings.TrimSpace(t.Title)
	}
	if len(desc) > 300 {
		desc = desc[:300] + "..."
	}

	var b strings.Builder
	b.WriteString("SnackVideo\n\n")
	if t.Creator.Name != "" {
		b.WriteString(fmt.Sprintf("Creator: %s\n", t.Creator.Name))
	}
	if t.Duration != "" {
		b.WriteString(fmt.Sprintf("Duration: %s\n", t.Duration))
	}
	if t.Interaction.Views > 0 || t.Interaction.Likes > 0 {
		b.WriteString(fmt.Sprintf("Stats: %s views · %s likes · %s shares\n",
			fmtCount(t.Interaction.Views),
			fmtCount(t.Interaction.Likes),
			fmtCount(t.Interaction.Shares)))
	}
	if desc != "" {
		b.WriteString("\n")
		b.WriteString(desc)
	}
	return b.String()
}

func fmtCount(n uint64) string {
	switch {
	case n >= 1_000_000_000:
		return fmt.Sprintf("%.1fB", float64(n)/1_000_000_000)
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}
