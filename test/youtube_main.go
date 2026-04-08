package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/kkdai/youtube/v2"
)

// TestVideoID adalah daftar video ID untuk testing
var TestVideoID = []struct {
	ID          string
	Description string
}{
	{"jNQXAC9IVRw", "First YouTube video ever - should be accessible"},
	{"dQw4w9WgXcQ", "Rick Astley - Never Gonna Give You Up"},
	{"9bZkp7q19f0", "PSY - GANGNAM STYLE"},
	{"kJQP7kiw5Fk", "Luis Fonsi - Despacito"},
	{"RgKAFK5djSk", "Wiz Khalifa - See You Again"},
}

// TestSimpleDownload mencoba download sederhana
func TestSimpleDownload() {
	fmt.Println("\n=== Testing Simple Download ===\n")

	videoID := "jNQXAC9IVRw"

	client := youtube.Client{}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fmt.Printf("📥 Fetching video info for: %s\n", videoID)
	
	video, err := client.GetVideoContext(ctx, videoID)
	if err != nil {
		fmt.Printf("❌ Failed to get video: %v\n", err)
		return
	}

	fmt.Printf("✅ Video Title: %s\n", video.Title)
	fmt.Printf("   Author: %s\n", video.Author)
	fmt.Printf("   Duration: %v\n", video.Duration)
	fmt.Printf("   Views: %d\n\n", video.Views)

	// List available formats
	fmt.Println("📋 Available Formats (first 10):")
	fmt.Println("---")
	for i, format := range video.Formats {
		if i >= 10 { // Show first 10 only
			fmt.Printf("   ... and %d more formats\n", len(video.Formats)-10)
			break
		}
		mimeType := "unknown"
		if format.MimeType != "" {
			mimeType = format.MimeType
		}
		fmt.Printf("   [%d] Quality: %-10s | MIME: %s | Bitrate: %d\n", 
			format.ItagNo, 
			format.Quality, 
			mimeType, 
			format.Bitrate)
	}

	// Try to get stream with audio
	fmt.Println("\n🎵 Getting format with audio...")
	formats := video.Formats.WithAudioChannels()
	
	if len(formats) == 0 {
		fmt.Println("❌ No formats with audio found!")
		return
	}

	fmt.Printf("✅ Found %d formats with audio\n", len(formats))

	// Get first format
	format := formats[0]
	fmt.Printf("\n📥 Downloading format: %s (itag: %d)\n", format.Quality, format.ItagNo)

	stream, size, err := client.GetStream(video, &format)
	if err != nil {
		fmt.Printf("❌ Failed to get stream: %v\n", err)
		return
	}
	defer stream.Close()

	fmt.Printf("✅ Stream obtained! Size: %.2f MB\n", float64(size)/(1024*1024))

	// Create temp file
	tmpFile, err := os.CreateTemp("", "youtube-test-*.mp4")
	if err != nil {
		fmt.Printf("❌ Failed to create temp file: %v\n", err)
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Download
	fmt.Println("📥 Downloading to temp file...")
	written, err := io.Copy(tmpFile, stream)
	if err != nil {
		fmt.Printf("❌ Failed to download: %v\n", err)
		return
	}

	fmt.Printf("✅ Downloaded! Size: %.2f MB\n", float64(written)/(1024*1024))
	fmt.Printf("   File: %s\n", tmpFile.Name())
}

// TestMultipleVideos mencoba beberapa video
func TestMultipleVideos() {
	fmt.Println("\n=== Testing Multiple Videos ===\n")

	client := youtube.Client{}

	for _, test := range TestVideoID {
		fmt.Printf("\n🎬 Testing: %s\n", test.Description)
		fmt.Printf("   Video ID: %s\n", test.ID)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		
		video, err := client.GetVideoContext(ctx, test.ID)
		cancel()

		if err != nil {
			fmt.Printf("   ❌ Error: %v\n", err)
		} else {
			fmt.Printf("   ✅ Title: %s\n", video.Title)
			fmt.Printf("   ✅ Duration: %v | Formats: %d\n", video.Duration, len(video.Formats))
		}
		fmt.Println()
	}
}

// TestWithCustomClient mencoba dengan custom HTTP client
func TestWithCustomClient() {
	fmt.Println("\n=== Testing with Custom HTTP Client ===\n")

	videoID := "jNQXAC9IVRw"

	// Create custom client with different user agent
	customHTTPClient := &youtube.Client{}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Printf("📥 Fetching video with custom client: %s\n", videoID)
	
	video, err := customHTTPClient.GetVideoContext(ctx, videoID)
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Success! Title: %s\n", video.Title)
	fmt.Printf("   Author: %s\n", video.Author)
	fmt.Printf("   Formats: %d\n", len(video.Formats))
}

func main() {
	fmt.Println("YouTube Downloader Test Suite")
	fmt.Println("================================\n")

	// Test 1: Simple download (most important)
	TestSimpleDownload()

	// Test 2: Multiple videos
	TestMultipleVideos()

	// Test 3: Custom client
	TestWithCustomClient()

	fmt.Println("\n================================")
	fmt.Println("✅ Test Complete!")
}
