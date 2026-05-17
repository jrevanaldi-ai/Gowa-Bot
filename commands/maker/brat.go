package maker

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

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

	author := strings.TrimSpace(ctx.PushName)
	if author == "" {
		author = ctx.Sender.User
	}

	stickerMsg, err := helper.BuildStickerFromImage(
		ctx.Ctx,
		ctx.Client,
		imageData,
		"Gowa-Bot",
		"@"+author,
		ctx.MessageID,
		ctx.Sender.String(),
		ctx.Chat.String(),
	)
	if err != nil {
		errorMsg := "Gagal konversi sticker.\n\n" +
			fmt.Sprintf("Error: %s", err.Error())
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
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
