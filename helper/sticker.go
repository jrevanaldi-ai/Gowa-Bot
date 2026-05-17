package helper

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"github.com/jrevanaldi-ai/gowa"
	"github.com/jrevanaldi-ai/gowa/proto/waE2E"
)

func ConvertImageToWebpSticker(input []byte) ([]byte, error) {
	if out, err := runImageMagick("magick", input); err == nil {
		return out, nil
	}
	if out, err := runImageMagick("convert", input); err == nil {
		return out, nil
	}
	if out, err := runFFmpeg(input); err == nil {
		return out, nil
	}
	return nil, fmt.Errorf("no image converter available — install ImageMagick (`apt install imagemagick`) or ffmpeg (`apt install ffmpeg`)")
}

func runImageMagick(bin string, input []byte) ([]byte, error) {
	args := []string{}
	if bin == "magick" {
		args = append(args, "convert")
	}
	args = append(args,
		"-",
		"-resize", "512x512",
		"-background", "none",
		"-gravity", "center",
		"-extent", "512x512",
		"-define", "webp:lossless=false",
		"-quality", "80",
		"webp:-",
	)
	cmd := exec.Command(bin, args...)
	cmd.Stdin = bytes.NewReader(input)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	if out.Len() == 0 {
		return nil, fmt.Errorf("%s produced empty output", bin)
	}
	return out.Bytes(), nil
}

func runFFmpeg(input []byte) ([]byte, error) {
	cmd := exec.Command("ffmpeg",
		"-hide_banner", "-loglevel", "error",
		"-i", "pipe:0",
		"-vf", "scale=512:512:force_original_aspect_ratio=decrease,pad=512:512:(ow-iw)/2:(oh-ih)/2:color=0x00000000",
		"-vcodec", "libwebp",
		"-lossless", "0",
		"-q:v", "80",
		"-preset", "default",
		"-an", "-vsync", "0",
		"-f", "webp",
		"pipe:1",
	)
	cmd.Stdin = bytes.NewReader(input)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	if out.Len() == 0 {
		return nil, fmt.Errorf("ffmpeg produced empty output")
	}
	return out.Bytes(), nil
}

func AddStickerMetadata(webp []byte, packName, packPublisher string) ([]byte, error) {
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

func BuildStickerFromImage(ctx context.Context, cli *gowa.Client, imgData []byte, packName, packPublisher, replyToMsgID, senderJID, chatJID string) (*waE2E.Message, error) {
	webpData, err := ConvertImageToWebpSticker(imgData)
	if err != nil {
		return nil, fmt.Errorf("convert sticker: %w", err)
	}

	if stamped, err := AddStickerMetadata(webpData, packName, packPublisher); err == nil {
		webpData = stamped
	}

	uploadResp, err := cli.Upload(ctx, webpData, gowa.MediaImage)
	if err != nil {
		return nil, fmt.Errorf("upload sticker: %w", err)
	}

	remoteJID := chatJID
	if remoteJID == "" {
		remoteJID = senderJID
	}

	return &waE2E.Message{
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
				StanzaID:    proto.String(replyToMsgID),
				Participant: proto.String(senderJID),
				RemoteJID:   proto.String(remoteJID),
			},
		},
	}, nil
}
