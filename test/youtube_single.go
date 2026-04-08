package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/kkdai/youtube/v2"
)

func main() {
	// Video ID yang benar dari URL: https://youtu.be/BjNPVbrsed0?si=YC_7ma55psMA-_-B
	videoID := "BjNPVbrsed0"

	fmt.Println("=== Testing Specific Video ===\n")
	fmt.Printf("🎬 Video ID: %s\n\n", videoID)

	client := youtube.Client{}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fmt.Println("📥 Fetching video info...")
	
	video, err := client.GetVideoContext(ctx, videoID)
	if err != nil {
		fmt.Printf("❌ Failed to get video: %v\n", err)
		
		// Try alternative approach
		fmt.Println("\n🔄 Trying alternative approach...")
		
		customClient := &youtube.Client{}
		ctx2, cancel2 := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel2()
		
		video, err = customClient.GetVideoContext(ctx2, videoID)
		if err != nil {
			fmt.Printf("❌ Still failed: %v\n", err)
			return
		}
	}

	fmt.Printf("✅ Video Title: %s\n", video.Title)
	fmt.Printf("   Author: %s\n", video.Author)
	fmt.Printf("   Duration: %v\n", video.Duration)
	fmt.Printf("   Views: %d\n", video.Views)
	fmt.Printf("   Formats available: %d\n\n", len(video.Formats))

	// List formats
	fmt.Println("📋 Available Formats (first 15):")
	fmt.Println("---")
	for i, format := range video.Formats {
		if i >= 15 {
			fmt.Printf("   ... and %d more\n", len(video.Formats)-15)
			break
		}
		mimeType := "unknown"
		if format.MimeType != "" {
			mimeType = format.MimeType
		}
		fmt.Printf("   [%d] Quality: %-12s | MIME: %-30s | Bitrate: %d\n", 
			format.ItagNo, 
			format.Quality, 
			mimeType, 
			format.Bitrate)
	}

	// Try download
	fmt.Println("\n🎵 Getting format with audio...")
	formats := video.Formats.WithAudioChannels()
	
	if len(formats) == 0 {
		fmt.Println("❌ No formats with audio found!")
		return
	}

	fmt.Printf("✅ Found %d formats with audio\n", len(formats))

	format := formats[0]
	fmt.Printf("\n📥 Attempting download: %s (itag: %d)\n", format.Quality, format.ItagNo)

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

	fmt.Println("📥 Downloading to temp file...")
	written, err := io.Copy(tmpFile, stream)
	if err != nil {
		fmt.Printf("❌ Failed to download: %v\n", err)
		return
	}

	fmt.Printf("\n✅✅✅ DOWNLOAD SUCCESS! ✅✅✅\n")
	fmt.Printf("   File size: %.2f MB\n", float64(written)/(1024*1024))
	fmt.Printf("   File: %s\n", tmpFile.Name())
}
