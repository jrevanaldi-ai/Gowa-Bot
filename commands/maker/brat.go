package maker

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"github.com/jrevanaldi-ai/gowa"
	"github.com/jrevanaldi-ai/gowa/proto/waE2E"
	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

var BratMetadata = &lib.CommandMetadata{
	Cmd:       "brat",
	Tag:       "maker",
	Desc:      "Buat sticker brat dari teks",
	Example:   ".brat halo dunia",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{},
}

func BratHandler(ctx *lib.CommandContext) error {
	text := strings.TrimSpace(strings.Join(ctx.Args, " "))

	if text == "" {
		message := "Masukkan teks untuk sticker brat.\n\n" +
			"Usage:\n" +
			"- .brat <teks>\n\n" +
			"Contoh:\n" +
			"- .brat halo dunia"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	apiURL := "https://api.azbry.com/api/maker/brat?text=" + url.QueryEscape(text)

	imageData, err := fetchBratImage(apiURL)
	if err != nil {
		errorMsg := "Gagal generate sticker.\n\n" +
			fmt.Sprintf("Error: %s", err.Error())
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	webpData, err := convertToWebpSticker(imageData)
	if err != nil {
		errorMsg := "Gagal konversi sticker.\n\n" +
			fmt.Sprintf("Error: %s", err.Error())
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	author := strings.TrimSpace(ctx.PushName)
	if author == "" {
		author = ctx.Sender.User
	}
	if stamped, err := addStickerMetadata(webpData, "Gowa-Bot", "@"+author); err == nil {
		webpData = stamped
	}

	uploadResp, err := ctx.Client.Upload(context.Background(), webpData, gowa.MediaImage)
	if err != nil {
		return fmt.Errorf("failed to upload sticker: %w", err)
	}

	stickerMsg := &waE2E.Message{
		StickerMessage: &waE2E.StickerMessage{
			URL:               proto.String(uploadResp.URL),
			DirectPath:        proto.String(uploadResp.DirectPath),
			Mimetype:          proto.String("image/webp"),
			FileSHA256:        uploadResp.FileSHA256,
			FileEncSHA256:     uploadResp.FileEncSHA256,
			FileLength:        proto.Uint64(uploadResp.FileLength),
			MediaKey:          uploadResp.MediaKey,
			MediaKeyTimestamp: proto.Int64(time.Now().Unix()),
			Height:            proto.Uint32(512),
			Width:             proto.Uint32(512),
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:    proto.String(ctx.MessageID),
				Participant: proto.String(ctx.Sender.String()),
				RemoteJID:   proto.String(ctx.Chat.String()),
			},
		},
	}

	if _, err := ctx.SendMessage(stickerMsg); err != nil {
		return fmt.Errorf("failed to send sticker: %w", err)
	}

	return nil
}

func fetchBratImage(apiURL string) ([]byte, error) {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "image/*,*/*;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	return data, nil
}

func addStickerMetadata(webp []byte, packName, packPublisher string) ([]byte, error) {
	if len(webp) < 12 || string(webp[0:4]) != "RIFF" || string(webp[8:12]) != "WEBP" {
		return nil, fmt.Errorf("not a valid WebP file")
	}

	metadata := struct {
		StickerPackID        string   `json:"sticker-pack-id"`
		StickerPackName      string   `json:"sticker-pack-name"`
		StickerPackPublisher string   `json:"sticker-pack-publisher"`
		Emojis               []string `json:"emojis"`
	}{
		StickerPackID:        uuid.New().String(),
		StickerPackName:      packName,
		StickerPackPublisher: packPublisher,
		Emojis:               []string{},
	}
	jsonBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	exifPayload := []byte{
		0x49, 0x49, 0x2A, 0x00,
		0x08, 0x00, 0x00, 0x00,
		0x01, 0x00,
		0x41, 0x57,
		0x07, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x16, 0x00, 0x00, 0x00,
	}
	binary.LittleEndian.PutUint32(exifPayload[14:18], uint32(len(jsonBytes)))
	exifPayload = append(exifPayload, jsonBytes...)

	var vp8x []byte
	var imageChunks []byte
	chunks := webp[12:]
	for pos := 0; pos+8 <= len(chunks); {
		fourCC := string(chunks[pos : pos+4])
		size := binary.LittleEndian.Uint32(chunks[pos+4 : pos+8])
		end := pos + 8 + int(size)
		if size%2 != 0 {
			end++
		}
		if end > len(chunks) {
			return nil, fmt.Errorf("malformed chunk %s", fourCC)
		}
		chunk := chunks[pos:end]
		switch fourCC {
		case "VP8X":
			vp8x = append([]byte{}, chunk...)
		case "EXIF", "XMP ":
		default:
			imageChunks = append(imageChunks, chunk...)
		}
		pos = end
	}

	if vp8x == nil {
		vp8x = []byte{
			'V', 'P', '8', 'X',
			0x0A, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0xFF, 0x01, 0x00,
			0xFF, 0x01, 0x00,
		}
	}
	vp8x[8] |= 0x08

	exifChunk := []byte{'E', 'X', 'I', 'F', 0, 0, 0, 0}
	binary.LittleEndian.PutUint32(exifChunk[4:8], uint32(len(exifPayload)))
	exifChunk = append(exifChunk, exifPayload...)
	if len(exifPayload)%2 != 0 {
		exifChunk = append(exifChunk, 0)
	}

	body := append([]byte{}, vp8x...)
	body = append(body, imageChunks...)
	body = append(body, exifChunk...)

	out := []byte{'R', 'I', 'F', 'F', 0, 0, 0, 0, 'W', 'E', 'B', 'P'}
	binary.LittleEndian.PutUint32(out[4:8], uint32(4+len(body)))
	out = append(out, body...)

	return out, nil
}

func convertToWebpSticker(input []byte) ([]byte, error) {
	cmd := exec.Command("convert",
		"-",
		"-resize", "512x512",
		"-background", "none",
		"-gravity", "center",
		"-extent", "512x512",
		"-define", "webp:lossless=false",
		"-quality", "80",
		"webp:-",
	)
	cmd.Stdin = bytes.NewReader(input)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	if out.Len() == 0 {
		return nil, fmt.Errorf("convert produced empty output")
	}
	return out.Bytes(), nil
}
