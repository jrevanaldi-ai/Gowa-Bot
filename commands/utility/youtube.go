package utility

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/jrevanaldi-ai/gowa"
	"github.com/jrevanaldi-ai/gowa/proto/waE2E"
	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

// YoutubeMetadata adalah metadata untuk command youtube
var YoutubeMetadata = &lib.CommandMetadata{
	Cmd:       "youtube",
	Tag:       "utility",
	Desc:      "Download video dari YouTube menggunakan yt-dlp (support age-restricted)",
	Example:   ".yt https://youtu.be/dQw4w9WgXcQ atau .yt -a <url> untuk audio saja",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"yt", "ytdl", "play"},
}

// YoutubeArgs argument untuk youtube command
type YoutubeArgs struct {
	URL       string
	AudioOnly bool
	Quality   string
}

// YtDlpInfo struktur untuk metadata video dari yt-dlp
type YtDlpInfo struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Author      string  `json:"channel"`
	Duration    float64 `json:"duration"`
	ViewCount   int64   `json:"view_count"`
	Description string  `json:"description"`
	Thumbnail   string  `json:"thumbnail"`
	Formats     []struct {
		FormatID   string  `json:"format_id"`
		FormatNote string  `json:"format_note"`
		Ext        string  `json:"ext"`
		ACodec     string  `json:"acodec"`
		VCodec     string  `json:"vcodec"`
		FileSize   float64 `json:"filesize"`
		Width      int     `json:"width"`
		Height     int     `json:"height"`
		FPS        float64 `json:"fps"`
		TBR        float64 `json:"tbr"`
	} `json:"formats"`
}

// YoutubeHandler menangani command youtube download
func YoutubeHandler(ctx *lib.CommandContext) error {
	// Cek apakah ada argument
	if len(ctx.Args) == 0 {
		return showYoutubeHelp(ctx)
	}

	// Parse arguments
	args := parseYoutubeArgs(ctx.Args)

	// Cek flag -help
	if args.URL == "help" || args.URL == "-h" || args.URL == "--help" {
		return showYoutubeHelp(ctx)
	}

	// Validasi URL
	if args.URL == "" {
		message := "❌ URL YouTube tidak boleh kosong!\n\n" +
			"Contoh:\n" +
			"• `.yt https://youtu.be/dQw4w9WgXcQ`\n" +
			"• `.yt https://www.youtube.com/watch?v=dQw4w9WgXcQ`"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Ekstrak video ID
	videoID, err := extractVideoID(args.URL)
	if err != nil {
		message := "❌ URL YouTube tidak valid!\n\n" +
			"Pastikan URL dalam format:\n" +
			"• `https://youtu.be/VIDEO_ID`\n" +
			"• `https://www.youtube.com/watch?v=VIDEO_ID`\n" +
			"• `https://www.youtube.com/shorts/VIDEO_ID`"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Cek apakah yt-dlp terinstall
	if !isYtDlpInstalled() {
		message := "❌ *yt-dlp tidak ditemukan!*\n\n" +
			"Bot memerlukan yt-dlp untuk download video.\n\n" +
			"*Solusi:*\n" +
			"• Install yt-dlp: `pip install yt-dlp`\n" +
			"• Atau hubungi admin bot"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Kirim loading message
	loadingMsg := "🔄 *Fetching video info...*\n\n" +
		"┌─⦿ *Info*\n" +
		fmt.Sprintf("│ • *Video ID:* %s\n", videoID) +
		fmt.Sprintf("│ • *Mode:* %s\n", getYoutubeModeText(args)) +
		"└──────────────\n\n" +
		"_Mohon tunggu..._"

	_, err = ctx.SendMessage(helper.CreateSimpleReply(loadingMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	if err != nil {
		return fmt.Errorf("failed to send loading message: %w", err)
	}

	// Download video/audio menggunakan yt-dlp
	return downloadWithYtDlp(ctx, videoID, args)
}

// parseYoutubeArgs memparse arguments dari command
func parseYoutubeArgs(args []string) *YoutubeArgs {
	result := &YoutubeArgs{
		Quality: "best",
	}

	i := 0
	for i < len(args) {
		arg := args[i]

		// Cek flags
		if arg == "-a" || arg == "--audio" || arg == "-audio" {
			result.AudioOnly = true
			i++
			continue
		}

		if arg == "-q" || arg == "--quality" {
			if i+1 < len(args) {
				result.Quality = strings.ToLower(args[i+1])
				i += 2
				continue
			}
		}

		if arg == "-h" || arg == "--help" || arg == "-help" {
			result.URL = "help"
			i++
			continue
		}

		// Jika bukan flag, anggap sebagai URL
		if strings.HasPrefix(arg, "http") || strings.HasPrefix(arg, "youtu") {
			result.URL = arg
			i++
			continue
		}

		i++
	}

	// Jika URL tidak ditemukan di awal, cari di seluruh args
	if result.URL == "" || result.URL != "help" {
		for _, arg := range args {
			if strings.HasPrefix(arg, "http") || strings.HasPrefix(arg, "youtu") {
				result.URL = arg
				break
			}
		}
	}

	return result
}

// showYoutubeHelp menampilkan bantuan untuk youtube command
func showYoutubeHelp(ctx *lib.CommandContext) error {
	message := "*🎬 YouTube Downloader (yt-dlp)*\n\n" +
		"┌─⦿ *Usage*\n" +
		"│ • `.yt <url>` - Download video\n" +
		"│ • `.yt -a <url>` - Download audio saja\n" +
		"│ • `.yt -q <quality> <url>` - Pilih kualitas\n" +
		"│ • `.yt -h` - Tampilkan bantuan ini\n" +
		"└──────────────\n\n" +
		"*📋 Flags:*\n" +
		"• `-a, --audio` - Download audio saja (MP3/M4A)\n" +
		"• `-q, --quality` - Kualitas (144/360/480/720/1080/best)\n" +
		"• `-h, --help` - Tampilkan bantuan ini\n\n" +
		"*🎯 Quality Options:*\n" +
		"• `144` - 144p (terendah)\n" +
		"• `360` - 360p (rendah)\n" +
		"• `480` - 480p (sedang)\n" +
		"• `720` - 720p HD\n" +
		"• `1080` - 1080p Full HD\n" +
		"• `best` - Kualitas terbaik (default)\n\n" +
		"*📝 Contoh:*\n" +
		"• `.yt https://youtu.be/dQw4w9WgXcQ`\n" +
		"• `.yt -a https://www.youtube.com/watch?v=dQw4w9WgXcQ`\n" +
		"• `.yt -q 720 https://youtu.be/dQw4w9WgXcQ`\n\n" +
		"*⚠️ Notes:*\n" +
		"• Video max 100MB untuk WhatsApp\n" +
		"• Audio dikirim sebagai dokumen MP3/M4A\n" +
		"• Support age-restricted videos (tergantung region)"

	_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return err
}

// downloadWithYtDlp mendownload video menggunakan yt-dlp
func downloadWithYtDlp(ctx *lib.CommandContext, videoID string, args *YoutubeArgs) error {
	// Buat temp directory
	tempDir, err := os.MkdirTemp("", "youtube-dl-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir) // Cleanup setelah selesai

	// Get video info dulu
	videoInfo, err := getVideoInfo(videoID)
	if err != nil {
		return sendYoutubeError(ctx, err, videoID)
	}

	// Cek durasi
	duration := time.Duration(videoInfo.Duration * float64(time.Second))
	if duration > 30*time.Minute {
		message := "❌ *Video terlalu panjang!*\n\n" +
			"┌─⦿ *Limit*\n" +
			"│ • Max durasi: 30 menit\n" +
			fmt.Sprintf("│ • Video ini: %s\n", formatDuration(duration)) +
			"└──────────────\n\n" +
			"Silakan pilih video yang lebih pendek."
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	// Format download command
	outputPath := filepath.Join(tempDir, "%(title)s.%(ext)s")
	
	ytDlpArgs := []string{
		"--no-playlist",
		"--no-warnings",
		"--no-check-certificates",
		"--no-mtime",
		"--restrict-filenames",
		"-o", outputPath,
	}

	// Tambahkan format/quality
	if args.AudioOnly {
		// Audio only
		ytDlpArgs = append(ytDlpArgs,
			"-x",
			"--audio-format", "m4a",
			"--audio-quality", "0",
		)
	} else {
		// Video dengan audio
		quality := args.Quality
		if quality == "" || quality == "best" {
			ytDlpArgs = append(ytDlpArgs,
				"-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best",
				"--merge-output-format", "mp4",
			)
		} else {
			ytDlpArgs = append(ytDlpArgs,
				"-f", fmt.Sprintf("bestvideo[height<=%s]+bestaudio/best[height<=%s]/best", quality, quality),
				"--merge-output-format", "mp4",
			)
		}
	}

	// Tambahkan URL
	ytDlpArgs = append(ytDlpArgs, args.URL)

	// Kirim download progress
	downloadMsg := fmt.Sprintf("📥 *Downloading...*\n\n"+
		"┌─⦿ *Video Info*\n"+
		"│ • *Title:* %s\n"+
		"│ • *Author:* %s\n"+
		"│ • *Duration:* %s\n"+
		"│ • *Views:* %s\n"+
		fmt.Sprintf("│ • *Mode:* %s\n", getYoutubeModeText(args)) +
		"└──────────────\n\n" +
		"_Mohon tunggu, sedang mendownload..._",
		truncateString(videoInfo.Title, 30),
		videoInfo.Author,
		formatDuration(duration),
		formatNumber(videoInfo.ViewCount))

	_, err = ctx.SendMessage(helper.CreateSimpleReply(downloadMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	if err != nil {
		return fmt.Errorf("failed to send download message: %w", err)
	}

	// Execute yt-dlp
	cmdCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "yt-dlp", ytDlpArgs...)
	cmd.Dir = tempDir
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("yt-dlp failed: %v\nOutput: %s", err, string(output))
	}

	// Cari file yang didownload
	files, err := os.ReadDir(tempDir)
	if err != nil {
		return fmt.Errorf("failed to read temp dir: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no file downloaded")
	}

	// Ambil file pertama
	downloadedFile := files[0]
	filePath := filepath.Join(tempDir, downloadedFile.Name())

	// Cek ukuran file
	fileInfo, err := downloadedFile.Info()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	maxSize := int64(100 * 1024 * 1024) // 100MB
	if fileInfo.Size() > maxSize {
		message := fmt.Sprintf("❌ *File terlalu besar!*\n\n"+
			"┌─⦿ *Info*\n"+
			"│ • Max size: 100 MB\n"+
			"│ • File size: %s\n"+
			"└──────────────\n\n"+
			"Gunakan kualitas lebih rendah:\n"+
			"• `.yt -q 480 <url>` untuk kualitas sedang\n"+
			"• `.yt -a <url>` untuk audio saja",
			helper.FormatFileSize(fileInfo.Size()))
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	// Upload dan kirim file
	return sendYoutubeFile(ctx, filePath, downloadedFile.Name(), videoInfo, duration, fileInfo.Size(), args.AudioOnly)
}

// getVideoInfo mendapatkan info video menggunakan yt-dlp
func getVideoInfo(videoID string) (*YtDlpInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "yt-dlp",
		"--dump-json",
		"--no-playlist",
		"--no-warnings",
		"--no-check-certificates",
		"-j",
		fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get video info: %v", err)
	}

	var info YtDlpInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, fmt.Errorf("failed to parse video info: %w", err)
	}

	return &info, nil
}

// sendYoutubeFile mengirim file yang sudah didownload
func sendYoutubeFile(ctx *lib.CommandContext, filePath, fileName string, videoInfo *YtDlpInfo, duration time.Duration, size int64, audioOnly bool) error {
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Kirim berdasarkan tipe
	if audioOnly {
		return sendYoutubeAudio(ctx, data, fileName, videoInfo, duration, size)
	}

	return sendYoutubeVideo(ctx, data, fileName, videoInfo, duration, size)
}

// sendYoutubeAudio mengirim file audio
func sendYoutubeAudio(ctx *lib.CommandContext, data []byte, fileName string, videoInfo *YtDlpInfo, duration time.Duration, size int64) error {
	// Upload audio
	resp, err := ctx.Client.Upload(context.Background(), data, gowa.MediaAudio)
	if err != nil {
		return fmt.Errorf("failed to upload audio: %w", err)
	}

	audioMsg := &waE2E.Message{
		AudioMessage: &waE2E.AudioMessage{
			URL:                 proto.String(resp.URL),
			DirectPath:          proto.String(resp.DirectPath),
			Mimetype:            proto.String("audio/mp4"),
			FileSHA256:          resp.FileSHA256,
			FileLength:          proto.Uint64(resp.FileLength),
			MediaKey:            resp.MediaKey,
			MediaKeyTimestamp:   proto.Int64(time.Now().Unix()),
			Seconds:             proto.Uint32(uint32(duration.Seconds())),
		},
	}

	// Audio messages tidak support reply di WhatsApp
	_, err = ctx.SendMessage(audioMsg)
	return err
}

// sendYoutubeVideo mengirim file video
func sendYoutubeVideo(ctx *lib.CommandContext, data []byte, fileName string, videoInfo *YtDlpInfo, duration time.Duration, size int64) error {
	// Upload video
	resp, err := ctx.Client.Upload(context.Background(), data, gowa.MediaVideo)
	if err != nil {
		return fmt.Errorf("failed to upload video: %w", err)
	}

	caption := fmt.Sprintf("*🎬 YouTube Video*\n\n"+
		"┌─⦿ *Info*\n"+
		"│ • *Title:* %s\n"+
		"│ • *Author:* %s\n"+
		"│ • *Duration:* %s\n"+
		"│ • *Size:* %s\n"+
		"│ • *Views:* %s\n"+
		"└──────────────\n\n"+
		"_Downloaded via Gowa-Bot (yt-dlp)_",
		videoInfo.Title,
		videoInfo.Author,
		formatDuration(duration),
		helper.FormatFileSize(size),
		formatNumber(videoInfo.ViewCount))

	videoMsg := &waE2E.Message{
		VideoMessage: &waE2E.VideoMessage{
			URL:                 proto.String(resp.URL),
			DirectPath:          proto.String(resp.DirectPath),
			Mimetype:            proto.String("video/mp4"),
			Caption:             proto.String(caption),
			FileSHA256:          resp.FileSHA256,
			FileLength:          proto.Uint64(resp.FileLength),
			MediaKey:            resp.MediaKey,
			MediaKeyTimestamp:   proto.Int64(time.Now().Unix()),
			Seconds:             proto.Uint32(uint32(duration.Seconds())),
		},
	}

	// Tambahkan reply context
	videoMsg = helper.BuildReplyMessage(videoMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String())

	_, err = ctx.SendMessage(videoMsg)
	return err
}

// isYtDlpInstalled cek apakah yt-dlp terinstall
func isYtDlpInstalled() bool {
	cmd := exec.Command("yt-dlp", "--version")
	err := cmd.Run()
	return err == nil
}

// extractVideoID mengekstrak video ID dari URL YouTube
func extractVideoID(url string) (string, error) {
	// Handle youtu.be short URL
	if strings.Contains(url, "youtu.be/") {
		parts := strings.Split(url, "youtu.be/")
		if len(parts) > 1 {
			videoID := strings.Split(parts[1], "?")[0]
			videoID = strings.Split(videoID, "&")[0]
			return videoID, nil
		}
	}

	// Handle youtube.com/watch?v=
	if strings.Contains(url, "youtube.com/watch") {
		re := regexp.MustCompile(`[?&]v=([^&]+)`)
		matches := re.FindStringSubmatch(url)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	// Handle youtube.com/shorts/
	if strings.Contains(url, "youtube.com/shorts/") {
		parts := strings.Split(url, "shorts/")
		if len(parts) > 1 {
			videoID := strings.Split(parts[1], "?")[0]
			videoID = strings.Split(videoID, "&")[0]
			return videoID, nil
		}
	}

	// Handle youtube.com/embed/
	if strings.Contains(url, "youtube.com/embed/") {
		parts := strings.Split(url, "embed/")
		if len(parts) > 1 {
			videoID := strings.Split(parts[1], "?")[0]
			videoID = strings.Split(videoID, "&")[0]
			return videoID, nil
		}
	}

	// Jika tidak match, coba extract langsung video ID
	if len(url) >= 11 && len(url) <= 20 && !strings.Contains(url, ".") {
		return url, nil
	}

	return "", fmt.Errorf("invalid YouTube URL")
}

// getYoutubeModeText mendapatkan text untuk mode download youtube
func getYoutubeModeText(args *YoutubeArgs) string {
	if args.AudioOnly {
		return "Audio Only 🎵"
	}
	return "Video + Audio 🎞️"
}

// formatDuration memformat durasi menjadi string yang readable
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// formatNumber memformat angka dengan separator
func formatNumber(n int64) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	} else if n >= 1000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

// sendYoutubeError mengirim pesan error yang user-friendly
func sendYoutubeError(ctx *lib.CommandContext, err error, videoID string) error {
	errMsg := err.Error()
	
	var message string
	
	switch {
	case strings.Contains(errMsg, "video is unavailable") || strings.Contains(errMsg, "not found"):
		message = "❌ *Video Tidak Ditemukan*\n\n" +
			"┌─⦿ *Error*\n" +
			"│ • Video ID: " + videoID + "\n" +
			"│ • Video mungkin sudah dihapus\n" +
			"└──────────────\n\n" +
			"*Solusi:*\n" +
			"• Cek kembali URL video\n" +
			"• Pastikan video masih tersedia"
			
	case strings.Contains(errMsg, "login required") || strings.Contains(errMsg, "bot"):
		message = "❌ *Verifikasi Diperlukan*\n\n" +
			"┌─⦿ *Error*\n" +
			"│ • YouTube memerlukan verifikasi\n" +
			"│ • Video dibatasi untuk pengguna tertentu\n" +
			"└──────────────\n\n" +
			"*Solusi:*\n" +
			"• Coba video lain\n" +
			"• Video ini mungkin terlalu dibatasi"
			
	case strings.Contains(errMsg, "private"):
		message = "❌ *Video Private*\n\n" +
			"┌─⦿ *Error*\n" +
			"│ • Video ini private\n" +
			"└──────────────\n\n" +
			"*Solusi:*\n" +
			"• Pilih video yang public"
			
	default:
		message = "❌ *Gagal Mendownload Video*\n\n" +
			"┌─⦿ *Error*\n" +
			"│ • " + truncateString(errMsg, 150) + "\n" +
			"└──────────────\n\n" +
			"*Solusi:*\n" +
			"• Coba video lain\n" +
			"• Pastikan URL valid"
	}
	
	_, sendErr := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	if sendErr != nil {
		return fmt.Errorf("failed to send error message: %w", sendErr)
	}
	
	return nil
}
