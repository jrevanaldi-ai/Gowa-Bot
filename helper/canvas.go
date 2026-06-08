package helper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type CanvasParams struct {
	IsWelcome     bool
	UserName      string
	GroupName     string
	MemberCount   int
	PfpURL        string
	GroupIconURL  string
	BackgroundURL string
}

const (
	siputzxBase           = "https://api.siputzx.my.id/api/canvas"
	defaultAvatarFallback = "https://files.catbox.moe/1xnz38.jpg"
)

func GenerateWelcomeCanvas(params CanvasParams) ([]byte, error) {
	endpoint := siputzxBase + "/welcomev1"
	if !params.IsWelcome {
		endpoint = siputzxBase + "/goodbyev1"
	}

	avatar := params.PfpURL
	if avatar == "" {
		avatar = defaultAvatarFallback
	}
	guildIcon := params.GroupIconURL
	if guildIcon == "" {
		guildIcon = defaultAvatarFallback
	}
	background := params.BackgroundURL
	if background == "" {
		background = defaultAvatarFallback
	}

	username := params.UserName
	if username == "" {
		username = "Member Baru"
	}
	guildName := params.GroupName
	if guildName == "" {
		guildName = "Group"
	}

	q := url.Values{}
	q.Set("username", username)
	q.Set("guildName", guildName)
	q.Set("memberCount", fmt.Sprintf("%d", params.MemberCount))
	q.Set("guildIcon", guildIcon)
	q.Set("avatar", avatar)
	q.Set("background", background)

	reqURL := endpoint + "?" + q.Encode()

	logger := NewLogger("Canvas")
	logger.Info("Canvas request: endpoint=%s user=%q group=%q count=%d", endpoint, username, guildName, params.MemberCount)
	logger.Info("Canvas avatar=%s", avatar)
	logger.Info("Canvas guildIcon=%s", guildIcon)
	logger.Info("Canvas background=%s", background)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "Gowa-Bot/1.0")
	req.Header.Set("Accept", "image/*")

	client := &http.Client{Timeout: 25 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call siputzx canvas: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("siputzx canvas %s: %s", resp.Status, string(body))
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "" && ct[:5] != "image" {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("siputzx returned non-image (%s): %s", ct, string(body))
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 8*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("siputzx returned empty body")
	}
	return data, nil
}
