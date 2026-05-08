package download

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/jrevanaldi-ai/gowa"
	"github.com/jrevanaldi-ai/gowa/proto/waE2E"
	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

var GitHubMetadata = &lib.CommandMetadata{
	Cmd:       "github",
	Tag:       "download",
	Desc:      "Download repository GitHub sebagai ZIP",
	Example:   ".github https://github.com/username/repo",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"gh", "gitclone", "git"},
}

type GitHubRepoResponse struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	HTMLURL     string `json:"html_url"`
	Size        int    `json:"size"`
	Stargazers  int    `json:"stargazers_count"`
	Forks       int    `json:"forks_count"`
	Language    string `json:"language"`
	Owner       struct {
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	} `json:"owner"`
}

func GitHubHandler(ctx *lib.CommandContext) error {
	if len(ctx.Args) == 0 {
		message := "❌ *Masukkan link GitHub!*\n\n" +
			"┌─⦿ *Usage*\n" +
			"│ • `.github <url>` - Download repo ZIP\n" +
			"└──────────────\n\n" +
			"*📝 Contoh:*\n" +
			"• `.github https://github.com/jrevanaldi-ai/gowa-bot`"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	githubURL := strings.Join(ctx.Args, " ")
	
	// Basic parsing: github.com/user/repo
	parts := strings.Split(strings.TrimPrefix(strings.TrimSuffix(githubURL, "/"), "https://"), "/")
	if len(parts) < 3 || !strings.Contains(parts[0], "github.com") {
		message := "❌ *URL GitHub tidak valid!*\n\n" +
			"┌─⦿ *Info*\n" +
			"│ • Format: `https://github.com/user/repo`\n" +
			"└──────────────"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	user := parts[1]
	repo := parts[2]

	// 1. Fetch Repo Info
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s", user, repo)
	repoInfo, err := fetchGitHubRepoInfo(apiURL)
	if err != nil {
		errorMsg := fmt.Sprintf("❌ *Gagal mengambil info repo!*\n\n│ • %s", err.Error())
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	// 2. Download ZIP
	zipURL := fmt.Sprintf("https://github.com/%s/%s/archive/refs/heads/main.zip", user, repo)
	// Try main branch first, then master if fail
	zipData, err := downloadFileGitHub(zipURL)
	if err != nil {
		// Try master
		zipURL = fmt.Sprintf("https://github.com/%s/%s/archive/refs/heads/master.zip", user, repo)
		zipData, err = downloadFileGitHub(zipURL)
		if err != nil {
			errorMsg := "❌ *Gagal mendownload ZIP!*\n\n│ • Pastikan branch utama adalah 'main' atau 'master'"
			_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
			return nil
		}
	}

	// 3. Upload to WhatsApp
	uploadResp, err := ctx.Client.Upload(context.Background(), zipData, gowa.MediaDocument)
	if err != nil {
		return fmt.Errorf("failed to upload zip: %w", err)
	}

	fileName := fmt.Sprintf("%s.zip", repo)
	caption := fmt.Sprintf("📦 *GitHub Repository*\n\n"+
		"┌─⦿ *Info*\n"+
		"│ • *Nama:* %s\n"+
		"│ • *Owner:* %s\n"+
		"│ • *Bintang:* %d ⭐\n"+
		"│ • *Forks:* %d 🍴\n"+
		"│ • *Bahasa:* %s\n"+
		"└──────────────\n\n"+
		"*Deskripsi:* %s", 
		repoInfo.Name, repoInfo.Owner.Login, repoInfo.Stargazers, repoInfo.Forks, repoInfo.Language, repoInfo.Description)

	senderStr := ctx.Sender.String()
	docMsg := &waE2E.Message{
		DocumentMessage: &waE2E.DocumentMessage{
			URL:           proto.String(uploadResp.URL),
			DirectPath:    proto.String(uploadResp.DirectPath),
			Mimetype:      proto.String("application/zip"),
			Title:         proto.String(fileName),
			FileName:      proto.String(fileName),
			FileSHA256:    uploadResp.FileSHA256,
			FileEncSHA256: uploadResp.FileEncSHA256,
			FileLength:    proto.Uint64(uploadResp.FileLength),
			MediaKey:      uploadResp.MediaKey,
			Caption:       proto.String(caption),
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:    &ctx.MessageID,
				Participant: &senderStr,
			},
		},
	}

	_, err = ctx.SendMessage(docMsg)
	return err
}

func fetchGitHubRepoInfo(apiURL string) (*GitHubRepoResponse, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}

	var repoInfo GitHubRepoResponse
	if err := json.NewDecoder(resp.Body).Decode(&repoInfo); err != nil {
		return nil, err
	}
	return &repoInfo, nil
}

func downloadFileGitHub(url string) ([]byte, error) {
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download error: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
