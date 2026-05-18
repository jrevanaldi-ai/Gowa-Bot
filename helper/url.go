package helper

import (
	"regexp"
	"strings"
)

var (
	whatsAppInviteRe = regexp.MustCompile(`(?i)^(?:https?://)?chat\.whatsapp\.com/(?:invite/)?([A-Za-z0-9_-]+)`)
	whatsAppCodeRe   = regexp.MustCompile(`^[A-Za-z0-9_-]{6,}$`)

	instagramRe = regexp.MustCompile(`(?i)^https?://(?:www\.)?instagram\.com/(?:p|reel|reels|tv|stories)/[A-Za-z0-9_-]+`)

	tiktokRe = regexp.MustCompile(`(?i)^https?://(?:www\.|vt\.|vm\.|m\.)?tiktok\.com/`)

	githubRe = regexp.MustCompile(`(?i)^https?://(?:www\.)?github\.com/([A-Za-z0-9._-]+)/([A-Za-z0-9._-]+?)(?:\.git)?(?:/.*)?$`)

	youtubeRe = regexp.MustCompile(`(?i)^https?://(?:www\.|m\.|music\.)?(?:youtube\.com/(?:watch\?v=|shorts/|embed/|v/|live/)|youtu\.be/)[A-Za-z0-9_-]+`)

	spotifyRe = regexp.MustCompile(`(?i)^https?://(?:open\.|www\.)?spotify\.com/(?:intl-[a-z]+/)?(?:track|album|playlist|artist|episode|show)/[A-Za-z0-9]+`)

	soundcloudRe = regexp.MustCompile(`(?i)^https?://(?:www\.|m\.|on\.)?soundcloud\.com/[A-Za-z0-9_-]+/[A-Za-z0-9_-]+`)

	snackvideoRe = regexp.MustCompile(`(?i)^https?://(?:www\.|m\.|s\.)?snackvideo\.com/`)

	urlSchemeRe = regexp.MustCompile(`(?i)^https?://`)

	anyURLRe = regexp.MustCompile(`(?i)https?://[^\s<>"'\x60]+`)
)

func ExtractFirstURL(text string) string {
	return anyURLRe.FindString(text)
}

func ExtractMatchingURL(text string, matcher func(string) bool) string {
	for _, u := range anyURLRe.FindAllString(text, -1) {
		if matcher(u) {
			return u
		}
	}
	return ""
}

func ExtractWhatsAppInviteCode(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	if m := whatsAppInviteRe.FindStringSubmatch(input); len(m) == 2 {
		return m[1]
	}

	if whatsAppCodeRe.MatchString(input) {
		return input
	}

	return ""
}

func IsInstagramURL(input string) bool {
	return instagramRe.MatchString(strings.TrimSpace(input))
}

func IsTikTokURL(input string) bool {
	return tiktokRe.MatchString(strings.TrimSpace(input))
}

func ExtractGitHubRepo(input string) (user, repo string, ok bool) {
	m := githubRe.FindStringSubmatch(strings.TrimSpace(input))
	if len(m) != 3 {
		return "", "", false
	}
	return m[1], m[2], true
}

func IsYouTubeURL(input string) bool {
	return youtubeRe.MatchString(strings.TrimSpace(input))
}

func IsSpotifyURL(input string) bool {
	return spotifyRe.MatchString(strings.TrimSpace(input))
}

func IsSoundCloudURL(input string) bool {
	return soundcloudRe.MatchString(strings.TrimSpace(input))
}

func IsSnackVideoURL(input string) bool {
	return snackvideoRe.MatchString(strings.TrimSpace(input))
}

func LooksLikeURL(input string) bool {
	return urlSchemeRe.MatchString(strings.TrimSpace(input))
}
