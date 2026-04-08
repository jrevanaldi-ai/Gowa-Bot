package utility

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	neturl "net/url"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/jrevanaldi-ai/gowa"
	"github.com/jrevanaldi-ai/gowa/proto/waE2E"
	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

// FetchMetadata adalah metadata untuk command fetch
var FetchMetadata = &lib.CommandMetadata{
	Cmd:       "fetch",
	Tag:       "utility",
	Desc:      "Melakukan HTTP request dan download media file",
	Example:   ".fetch https://api.github.com/users/octocat atau .fetch -media <url>",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"curl", "http"},
}

// FetchHandler menangani command fetch
func FetchHandler(ctx *lib.CommandContext) error {
	// Cek apakah ada argument
	if len(ctx.Args) == 0 {
		return showFetchHelp(ctx)
	}

	// Parse arguments
	args := parseFetchArgs(ctx.Args)

	// Cek flag -help
	if args.Help {
		return showFetchHelp(ctx)
	}

	// Validasi URL
	if args.URL == "" {
		message := "❌ URL tidak boleh kosong!\n\n" +
			"Contoh:\n" +
			"• `.fetch https://api.github.com`"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Validasi protocol URL
	if !strings.HasPrefix(args.URL, "http://") && !strings.HasPrefix(args.URL, "https://") {
		message := "❌ URL harus dimulai dengan http:// atau https://\n\n" +
			"Contoh:\n" +
			"• `.fetch https://api.github.com`"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Kirim request
	return executeFetchRequest(ctx, args)
}

// FetchArgs argument untuk fetch command
type FetchArgs struct {
	URL       string
	Method    string
	Headers   map[string]string
	Data      string
	Timeout   int
	Help      bool
	MediaMode bool // Jika true, force download media file
}

// parseFetchArgs memparse arguments dari command
func parseFetchArgs(args []string) *FetchArgs {
	result := &FetchArgs{
		Method:    "GET",
		Headers:   make(map[string]string),
		Timeout:   30,
		Help:      false,
		MediaMode: false,
	}

	i := 0
	for i < len(args) {
		arg := args[i]

		// Cek flags
		if arg == "-h" || arg == "--help" || arg == "-help" {
			result.Help = true
			i++
			continue
		}

		if arg == "-m" || arg == "--media" {
			result.MediaMode = true
			i++
			continue
		}

		if arg == "-X" || arg == "--method" {
			if i+1 < len(args) {
				result.Method = strings.ToUpper(args[i+1])
				i += 2
				continue
			}
		}

		if arg == "-H" || arg == "--header" {
			if i+1 < len(args) {
				header := args[i+1]
				parts := strings.SplitN(header, ":", 2)
				if len(parts) == 2 {
					result.Headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
				i += 2
				continue
			}
		}

		if arg == "-d" || arg == "--data" {
			if i+1 < len(args) {
				result.Data = args[i+1]
				result.Method = "POST" // Auto set to POST if data provided
				i += 2
				continue
			}
		}

		if arg == "-t" || arg == "--timeout" {
			if i+1 < len(args) {
				// Parse timeout (dalam detik)
				timeout := 30
				fmt.Sscanf(args[i+1], "%d", &timeout)
				result.Timeout = timeout
				i += 2
				continue
			}
		}

		// Jika bukan flag, anggap sebagai URL
		if strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://") {
			result.URL = arg
			i++
			continue
		}

		i++
	}

	return result
}

// showFetchHelp menampilkan bantuan untuk fetch command
func showFetchHelp(ctx *lib.CommandContext) error {
	message := "*🌐 Fetch Command Help*\n\n" +
		"┌─⦿ *Usage*\n" +
		"│ • `.fetch <url>` - GET request\n" +
		"│ • `.fetch -m <url>` - Download media file\n" +
		"│ • `.fetch -X POST <url>` - POST request\n" +
		"│ • `.fetch -H \"Content-Type: application/json\" <url>`\n" +
		"│ • `.fetch -d '{\"key\":\"value\"}' <url>`\n" +
		"│ • `.fetch -t 60 <url>` - Timeout 60 detik\n" +
		"│ • `.fetch -h` - Tampilkan bantuan ini\n" +
		"└──────────────\n\n" +
		"*📋 Flags:*\n" +
		"• `-X, --method` - HTTP method (GET, POST, PUT, DELETE, etc)\n" +
		"• `-H, --header` - Custom header (format: \"Key: Value\")\n" +
		"• `-d, --data` - Data untuk POST/PUT request\n" +
		"• `-t, --timeout` - Timeout dalam detik (default: 30)\n" +
		"• `-m, --media` - Force download sebagai media file\n" +
		"• `-h, --help` - Tampilkan bantuan ini\n\n" +
		"*📝 Contoh:*\n" +
		"• `.fetch https://api.github.com/users/octocat`\n" +
		"• `.fetch -m https://example.com/image.jpg`\n" +
		"• `.fetch -X POST https://httpbin.org/post -d '{\"name\":\"test\"}'`\n" +
		"• `.fetch -H \"Authorization: Bearer token\" https://api.example.com/data`\n\n" +
		"*🎬 Supported Media:*\n" +
		"• 🖼️ *Image:* JPG, PNG, GIF, WebP, BMP\n" +
		"• 🎵 *Audio:* MP3, WAV, OGG, AAC, M4A\n" +
		"• 🎥 *Video:* MP4, WebM, AVI, MKV\n" +
		"• 📄 *Document:* PDF, ZIP, TXT, DOC, XLS"

	_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return err
}

// executeFetchRequest mengeksekusi HTTP request
func executeFetchRequest(ctx *lib.CommandContext, args *FetchArgs) error {
	// Tampilkan loading message
	loadingMsg := "🔄 *Fetching...*\n\n" +
		"┌─⦿ *Request Info*\n" +
		fmt.Sprintf("│ • *Method:* %s\n", args.Method) +
		fmt.Sprintf("│ • *URL:* %s\n", truncateString(args.URL, 40)) +
		fmt.Sprintf("│ • *Timeout:* %d seconds\n", args.Timeout) +
		fmt.Sprintf("│ • *Mode:* %s\n", getModeText(args.MediaMode)) +
		"└──────────────\n\n" +
		"_Mohon tunggu..._"

	loadingResp, err := ctx.SendMessage(helper.CreateSimpleReply(loadingMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	if err != nil {
		return fmt.Errorf("failed to send loading message: %w", err)
	}
	_ = loadingResp // Ignore for now, can be used for message edit

	// Buat HTTP client dengan timeout
	client := &http.Client{
		Timeout: time.Duration(args.Timeout) * time.Second,
	}

	// Buat request
	var req *http.Request
	if args.Data != "" {
		req, err = http.NewRequest(args.Method, args.URL, strings.NewReader(args.Data))
	} else {
		req, err = http.NewRequest(args.Method, args.URL, nil)
	}

	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Gowa-Bot/1.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,id;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	for key, value := range args.Headers {
		req.Header.Set(key, value)
	}

	// Eksekusi request
	startTime := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(startTime).Milliseconds()

	if err != nil {
		// Handle error
		errorMsg := "❌ *Request Failed*\n\n" +
			"┌─⦿ *Error*\n" +
			fmt.Sprintf("│ • %s\n", err.Error()) +
			"└──────────────"

		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil // Don't return error, just show message
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Cek content-type dari header
	contentType := resp.Header.Get("Content-Type")
	contentLength := len(body)

	// Deteksi MIME type dari data jika content-type kosong atau tidak jelas
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = http.DetectContentType(body)
	}

	// Cek apakah URL memiliki ekstensi media
	ext := strings.ToLower(filepath.Ext(args.URL))
	isMediaExt := isMediaExtension(ext)

	// Jika media mode, content-type adalah media, atau URL memiliki ekstensi media, kirim sebagai file
	if args.MediaMode || isMediaFile(contentType, args.URL) || isMediaExt {
		return sendMediaFile(ctx, body, contentType, args.URL, latency, ext)
	}

	// Jika text dan kecil, tampilkan sebagai text
	if isTextFile(contentType, args.URL) && contentLength < 5000 {
		responseMsg := formatFetchResponse(resp, body, latency, args.URL)
		_, err = ctx.SendMessage(helper.CreateSimpleReply(responseMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Jika file terlalu besar untuk ditampilkan, tawarkan untuk download
	if contentLength > 5000 {
		fileName := getFileNameFromURL(args.URL)
		if fileName == "" {
			fileName = "download"
		}
		
		message := fmt.Sprintf("*📦 File Response*\n\n"+
			"┌─⦿ *Info*\n"+
			fmt.Sprintf("│ • *Size:* %s\n", formatFileSize(contentLength))+
			fmt.Sprintf("│ • *Type:* %s\n", contentType)+
			fmt.Sprintf("│ • *Latency:* %d ms\n", latency)+
			"└──────────────\n\n"+
			"Gunakan `.fetch -m %s` untuk download file ini.", truncateString(args.URL, 30))

		_, err = ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Default: tampilkan sebagai text
	responseMsg := formatFetchResponse(resp, body, latency, args.URL)
	_, err = ctx.SendMessage(helper.CreateSimpleReply(responseMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return err
}

// formatFetchResponse memformat response untuk ditampilkan
func formatFetchResponse(resp *http.Response, body []byte, latency int64, url string) string {
	// Truncate body jika terlalu panjang (max 2000 karakter)
	bodyStr := string(body)
	if len(bodyStr) > 2000 {
		bodyStr = bodyStr[:2000] + "\n\n... (truncated)"
	}

	// Format headers
	var headersBuilder strings.Builder
	for key, values := range resp.Header {
		for _, value := range values {
			headersBuilder.WriteString(fmt.Sprintf("│ • %s: %s\n", key, truncateString(value, 30)))
		}
	}

	// Parse URL untuk display
	parsedURL, _ := neturl.Parse(url)
	host := ""
	if parsedURL != nil {
		host = parsedURL.Host
	}

	message := fmt.Sprintf("*✅ Response from %s*\n\n", host) +
		"┌─⦿ *Status*\n" +
		fmt.Sprintf("│ • *Code:* %d %s\n", resp.StatusCode, getStatusText(resp.StatusCode)) +
		fmt.Sprintf("│ • *Protocol:* %s\n", resp.Proto) +
		fmt.Sprintf("│ • *Latency:* %d ms\n", latency) +
		"└──────────────\n\n" +
		"┌─⦿ *Headers*\n"

	if headersBuilder.Len() == 0 {
		message += "│ • (no headers)\n"
	} else {
		message += headersBuilder.String()
	}

	message += "└──────────────\n\n" +
		"┌─⦿ *Body*\n" +
		"│ ```\n"

	// Format body (truncate untuk preview)
	previewBody := bodyStr
	if len(previewBody) > 1500 {
		previewBody = previewBody[:1500] + "\n... (truncated, see full in next message)"
	}

	message += previewBody + "\n```" +
		"\n└──────────────"

	return message
}

// getStatusText mendapatkan text untuk status code
func getStatusText(code int) string {
	switch code {
	case 200:
		return "✓ OK"
	case 201:
		return "✓ Created"
	case 204:
		return "✓ No Content"
	case 301:
		return "→ Moved Permanently"
	case 302:
		return "→ Found"
	case 400:
		return "✗ Bad Request"
	case 401:
		return "✗ Unauthorized"
	case 403:
		return "✗ Forbidden"
	case 404:
		return "✗ Not Found"
	case 500:
		return "✗ Internal Server Error"
	case 502:
		return "✗ Bad Gateway"
	case 503:
		return "✗ Service Unavailable"
	default:
		return ""
	}
}

// truncateString memotong string jika terlalu panjang
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// getModeText mendapatkan text untuk mode
func getModeText(mediaMode bool) string {
	if mediaMode {
		return "Media Download 📥"
	}
	return "Text Response 📝"
}

// isMediaExtension cek apakah ekstensi adalah ekstensi media
func isMediaExtension(ext string) bool {
	mediaExtensions := []string{
		// Images
		".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".ico", ".svg",
		// Audio
		".mp3", ".wav", ".ogg", ".aac", ".m4a", ".flac", ".wma",
		// Video
		".mp4", ".webm", ".avi", ".mkv", ".mov", ".wmv", ".flv",
		// Documents
		".pdf", ".zip", ".rar", ".tar", ".gz", ".doc", ".docx", ".xls", ".xlsx",
	}
	
	for _, mediaExt := range mediaExtensions {
		if ext == mediaExt {
			return true
		}
	}
	
	return false
}

// isImageExtension cek apakah ekstensi adalah image
func isImageExtension(ext string) bool {
	imageExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".ico", ".svg"}
	for _, imgExt := range imageExtensions {
		if ext == imgExt {
			return true
		}
	}
	return false
}

// isAudioExtension cek apakah ekstensi adalah audio
func isAudioExtension(ext string) bool {
	audioExtensions := []string{".mp3", ".wav", ".ogg", ".aac", ".m4a", ".flac", ".wma"}
	for _, audExt := range audioExtensions {
		if ext == audExt {
			return true
		}
	}
	return false
}

// isVideoExtension cek apakah ekstensi adalah video
func isVideoExtension(ext string) bool {
	videoExtensions := []string{".mp4", ".webm", ".avi", ".mkv", ".mov", ".wmv", ".flv"}
	for _, vidExt := range videoExtensions {
		if ext == vidExt {
			return true
		}
	}
	return false
}

// isMediaFile cek apakah content-type atau URL mengindikasikan media file
func isMediaFile(contentType, url string) bool {
	// Cek content-type
	if isImageType(contentType) || isAudioType(contentType) || isVideoType(contentType) {
		return true
	}

	// Cek dari ekstensi file di URL
	ext := strings.ToLower(filepath.Ext(url))
	mediaExtensions := []string{
		// Images
		".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".ico", ".svg",
		// Audio
		".mp3", ".wav", ".ogg", ".aac", ".m4a", ".flac", ".wma",
		// Video
		".mp4", ".webm", ".avi", ".mkv", ".mov", ".wmv", ".flv",
		// Documents
		".pdf", ".zip", ".rar", ".tar", ".gz", ".doc", ".docx", ".xls", ".xlsx",
	}

	for _, mediaExt := range mediaExtensions {
		if ext == mediaExt {
			return true
		}
	}

	return false
}

// isImageType cek apakah content-type adalah image
func isImageType(contentType string) bool {
	return strings.HasPrefix(contentType, "image/")
}

// isAudioType cek apakah content-type adalah audio
func isAudioType(contentType string) bool {
	return strings.HasPrefix(contentType, "audio/")
}

// isVideoType cek apakah content-type adalah video
func isVideoType(contentType string) bool {
	return strings.HasPrefix(contentType, "video/")
}

// isTextFile cek apakah content-type adalah text
func isTextFile(contentType, url string) bool {
	if strings.HasPrefix(contentType, "text/") {
		return true
	}
	if strings.HasPrefix(contentType, "application/json") {
		return true
	}
	if strings.HasPrefix(contentType, "application/xml") {
		return true
	}

	// Cek dari ekstensi
	ext := strings.ToLower(filepath.Ext(url))
	textExtensions := []string{".txt", ".json", ".xml", ".md", ".html", ".css", ".js"}

	for _, textExt := range textExtensions {
		if ext == textExt {
			return true
		}
	}

	return false
}

// sendMediaFile mengirim file media ke WhatsApp
func sendMediaFile(ctx *lib.CommandContext, data []byte, contentType, url string, latency int64, urlExt string) error {
	fileName := getFileNameFromURL(url)
	if fileName == "" {
		fileName = "download"
	}

	// Gunakan ekstensi dari URL jika ada, jika tidak gunakan dari MIME type
	ext := urlExt
	if ext == "" || !isMediaExtension(ext) {
		ext = getExtensionFromMimeType(contentType)
	}
	
	if !hasExtension(fileName) && ext != "" {
		fileName += ext
	}

	// Deteksi ulang content-type dari data untuk memastikan akurasi
	detectedType := http.DetectContentType(data)
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = detectedType
	}

	// Kirim berdasarkan tipe media
	var err error
	switch {
	case isImageType(contentType) || isImageExtension(urlExt):
		err = sendImageMessage(ctx, data, fileName, latency)
	case isAudioType(contentType) || isAudioExtension(urlExt):
		err = sendAudioMessage(ctx, data, fileName, latency)
	case isVideoType(contentType) || isVideoExtension(urlExt):
		err = sendVideoMessage(ctx, data, fileName, latency)
	default:
		err = sendDocumentMessage(ctx, data, fileName, contentType, latency)
	}

	if err != nil {
		return fmt.Errorf("failed to send media: %w", err)
	}

	return nil
}

// sendImageMessage mengirim pesan gambar
func sendImageMessage(ctx *lib.CommandContext, data []byte, fileName string, latency int64) error {
	// Upload gambar
	resp, err := ctx.Client.Upload(context.Background(), data, gowa.MediaImage)
	if err != nil {
		return fmt.Errorf("failed to upload image: %w", err)
	}

	// Buat pesan gambar
	imageMsg := &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			URL:           proto.String(resp.URL),
			DirectPath:    proto.String(resp.DirectPath),
			Mimetype:      proto.String(http.DetectContentType(data)),
			Caption:       proto.String(fmt.Sprintf("*🖼️ Image*\n\n📁 *File:* %s\n⚡ *Latency:* %d ms", fileName, latency)),
			FileSHA256:    resp.FileSHA256,
			FileLength:    proto.Uint64(resp.FileLength),
			MediaKey:      resp.MediaKey,
			MediaKeyTimestamp: proto.Int64(time.Now().Unix()),
		},
	}

	// Tambahkan reply context
	imageMsg = helper.BuildReplyMessage(imageMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String())

	_, err = ctx.SendMessage(imageMsg)
	return err
}

// sendAudioMessage mengirim pesan audio
func sendAudioMessage(ctx *lib.CommandContext, data []byte, fileName string, latency int64) error {
	// Upload audio
	resp, err := ctx.Client.Upload(context.Background(), data, gowa.MediaAudio)
	if err != nil {
		return fmt.Errorf("failed to upload audio: %w", err)
	}

	// Buat pesan audio
	audioMsg := &waE2E.Message{
		AudioMessage: &waE2E.AudioMessage{
			URL:           proto.String(resp.URL),
			DirectPath:    proto.String(resp.DirectPath),
			Mimetype:      proto.String(http.DetectContentType(data)),
			FileSHA256:    resp.FileSHA256,
			FileLength:    proto.Uint64(resp.FileLength),
			MediaKey:      resp.MediaKey,
			MediaKeyTimestamp: proto.Int64(time.Now().Unix()),
			Seconds:       proto.Uint32(0),
		},
	}

	_, err = ctx.SendMessage(audioMsg)
	return err
}

// sendVideoMessage mengirim pesan video
func sendVideoMessage(ctx *lib.CommandContext, data []byte, fileName string, latency int64) error {
	// Upload video
	resp, err := ctx.Client.Upload(context.Background(), data, gowa.MediaVideo)
	if err != nil {
		return fmt.Errorf("failed to upload video: %w", err)
	}

	// Buat pesan video
	videoMsg := &waE2E.Message{
		VideoMessage: &waE2E.VideoMessage{
			URL:           proto.String(resp.URL),
			DirectPath:    proto.String(resp.DirectPath),
			Mimetype:      proto.String(http.DetectContentType(data)),
			Caption:       proto.String(fmt.Sprintf("*🎥 Video*\n\n📁 *File:* %s\n⚡ *Latency:* %d ms", fileName, latency)),
			FileSHA256:    resp.FileSHA256,
			FileLength:    proto.Uint64(resp.FileLength),
			MediaKey:      resp.MediaKey,
			MediaKeyTimestamp: proto.Int64(time.Now().Unix()),
			Seconds:       proto.Uint32(0),
		},
	}

	// Tambahkan reply context
	videoMsg = helper.BuildReplyMessage(videoMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String())

	_, err = ctx.SendMessage(videoMsg)
	return err
}

// sendDocumentMessage mengirim pesan dokumen
func sendDocumentMessage(ctx *lib.CommandContext, data []byte, fileName, contentType string, latency int64) error {
	// Upload dokumen
	resp, err := ctx.Client.Upload(context.Background(), data, gowa.MediaDocument)
	if err != nil {
		return fmt.Errorf("failed to upload document: %w", err)
	}

	// Buat pesan dokumen
	docMsg := &waE2E.Message{
		DocumentMessage: &waE2E.DocumentMessage{
			URL:           proto.String(resp.URL),
			DirectPath:    proto.String(resp.DirectPath),
			Mimetype:      proto.String(contentType),
			Title:         proto.String(fileName),
			FileName:      proto.String(fileName),
			FileSHA256:    resp.FileSHA256,
			FileLength:    proto.Uint64(resp.FileLength),
			MediaKey:      resp.MediaKey,
			MediaKeyTimestamp: proto.Int64(time.Now().Unix()),
		},
	}

	// Tambahkan reply context
	docMsg = helper.BuildReplyMessage(docMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String())

	_, err = ctx.SendMessage(docMsg)
	return err
}

// getFileNameFromURL mendapatkan nama file dari URL
func getFileNameFromURL(url string) string {
	parsed, err := neturl.Parse(url)
	if err != nil {
		return ""
	}

	path := parsed.Path
	if path == "" {
		return ""
	}

	// Dapatkan nama file dari path
	_, fileName := filepath.Split(path)
	return fileName
}

// getExtensionFromMimeType mendapatkan ekstensi file dari MIME type
func getExtensionFromMimeType(mimeType string) string {
	extensions, err := mime.ExtensionsByType(mimeType)
	if err != nil || len(extensions) == 0 {
		return ""
	}
	return extensions[0]
}

// hasExtension cek apakah filename sudah punya ekstensi
func hasExtension(fileName string) bool {
	ext := filepath.Ext(fileName)
	return ext != "" && ext != "."
}

// formatFileSize memformat ukuran file menjadi human readable
func formatFileSize(size int) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	if size < KB {
		return fmt.Sprintf("%d B", size)
	} else if size < MB {
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	} else if size < GB {
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	}
	return fmt.Sprintf("%.2f GB", float64(size)/GB)
}
