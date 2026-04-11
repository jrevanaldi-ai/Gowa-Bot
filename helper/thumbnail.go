package helper

import (
	"bytes"
	"fmt"
	"image/color"
	"math"

	"github.com/fogleman/gg"
)


type ThumbnailConfig struct {
	Title       string
	Subtitle    string
	Footer      string
	Icon        string
	Width       int
	Height      int
	BgColor1    color.Color
	BgColor2    color.Color
	AccentColor color.Color
}


func DefaultThumbnailConfig() *ThumbnailConfig {
	return &ThumbnailConfig{
		Title:       "GOWA-BOT",
		Subtitle:    "WhatsApp Bot",
		Footer:      "Powered by Gowa",
		Icon:        "🤖",
		Width:       800,
		Height:      400,
		BgColor1:    color.RGBA{20, 20, 60, 255},
		BgColor2:    color.RGBA{60, 20, 80, 255},
		AccentColor: color.RGBA{0, 200, 255, 255},
	}
}


func CreateThumbnail(config *ThumbnailConfig) ([]byte, error) {
	if config == nil {
		config = DefaultThumbnailConfig()
	}

	dc := gg.NewContext(config.Width, config.Height)


	drawGradient(dc, config.Width, config.Height, config.BgColor1, config.BgColor2)


	drawDecorations(dc, config.Width, config.Height, config.AccentColor)


	drawMainText(dc, config.Title, config.AccentColor)
	drawSubtitleText(dc, config.Subtitle, config.Width, config.Height)
	drawFooterText(dc, config.Footer, config.Width, config.Height)


	drawIcon(dc, config.Icon, config.Width, config.Height)


	var buf bytes.Buffer
	if err := dc.EncodePNG(&buf); err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %w", err)
	}

	return buf.Bytes(), nil
}


func drawGradient(dc *gg.Context, width, height int, color1, color2 color.Color) {
	for y := 0; y < height; y++ {
		progress := float64(y) / float64(height)
		r := interpolate(color1, color2, progress).(*color.RGBA)
		dc.SetRGB(float64(r.R)/255.0, float64(r.G)/255.0, float64(r.B)/255.0)
		dc.DrawLine(0, float64(y), float64(width), float64(y))
		dc.Stroke()
	}
}


func drawDecorations(dc *gg.Context, width, height int, accentColor color.Color) {

	accent := accentColor.(*color.RGBA)

	
	dc.SetRGBA(float64(accent.R)/255.0, float64(accent.G)/255.0, float64(accent.B)/255.0, 0.1)
	for i := 0; i < 5; i++ {
		x := float64(width) * 0.1 * float64(i+1)
		y := float64(height) * 0.5
		radius := 20.0 + float64(i)*10.0
		dc.DrawCircle(x, y, radius)
		dc.Fill()
	}

	
	dc.SetRGBA(1, 1, 1, 0.05)
	dc.DrawRectangle(0, 0, float64(width), float64(height))
	dc.Fill()

	
	dc.SetRGBA(float64(accent.R)/255.0, float64(accent.G)/255.0, float64(accent.B)/255.0, 0.3)
	dc.SetLineWidth(3)
	dc.DrawRectangle(10, 10, float64(width)-20, float64(height)-20)
	dc.Stroke()
}


func drawMainText(dc *gg.Context, text string, accentColor color.Color) {
	dc.LoadFontFace("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf", 48)
	accent := accentColor.(*color.RGBA)
	dc.SetRGBA(float64(accent.R)/255.0, float64(accent.G)/255.0, float64(accent.B)/255.0, 1.0)
	dc.DrawStringAnchored(text, 400, 150, 0.5, 0.5)
}


func drawSubtitleText(dc *gg.Context, text string, width, height int) {
	dc.LoadFontFace("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", 28)
	dc.SetRGBA(1, 1, 1, 0.8)
	dc.DrawStringAnchored(text, float64(width)/2, float64(height)/2+40, 0.5, 0.5)
}


func drawFooterText(dc *gg.Context, text string, width, height int) {
	dc.LoadFontFace("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", 18)
	dc.SetRGBA(1, 1, 1, 0.5)
	dc.DrawStringAnchored(text, float64(width)/2, float64(height)-40, 0.5, 0.5)
}


func drawIcon(dc *gg.Context, icon string, width, height int) {
	
	dc.LoadFontFace("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", 60)
	dc.SetRGBA(1, 1, 1, 0.2)
	dc.DrawStringAnchored(icon, float64(width)-80, 80, 0.5, 0.5)
}


func interpolate(c1, c2 color.Color, t float64) color.Color {
	rgba1 := c1.(*color.RGBA)
	rgba2 := c2.(*color.RGBA)
	
	r := uint8(float64(rgba1.R) + (float64(rgba2.R)-float64(rgba1.R))*t)
	g := uint8(float64(rgba1.G) + (float64(rgba2.G)-float64(rgba1.G))*t)
	b := uint8(float64(rgba1.B) + (float64(rgba2.B)-float64(rgba1.B))*t)
	a := uint8(float64(rgba1.A) + (float64(rgba2.A)-float64(rgba1.A))*t)
	
	return &color.RGBA{R: r, G: g, B: b, A: a}
}


func CreateWelcomeCard(name string, groupName string) ([]byte, error) {
	config := DefaultThumbnailConfig()
	config.Title = "👋 WELCOME"
	config.Subtitle = fmt.Sprintf("%s", name)
	config.Footer = groupName
	config.AccentColor = color.RGBA{0, 255, 100, 255}
	config.BgColor1 = color.RGBA{20, 60, 20, 255}
	config.BgColor2 = color.RGBA{20, 20, 60, 255}
	
	return CreateThumbnail(config)
}


func CreateMenuCard() ([]byte, error) {
	config := DefaultThumbnailConfig()
	config.Title = "📋 MENU"
	config.Subtitle = "GOWA-BOT Commands"
	config.Footer = "Use .help <command> for details"
	
	return CreateThumbnail(config)
}


func CreateInfoCard(title string, message string) ([]byte, error) {
	config := DefaultThumbnailConfig()
	config.Title = title
	config.Subtitle = message
	config.Footer = "GOWA-BOT"
	config.AccentColor = color.RGBA{255, 200, 0, 255}
	config.BgColor1 = color.RGBA{60, 40, 20, 255}
	config.BgColor2 = color.RGBA{40, 20, 60, 255}
	
	return CreateThumbnail(config)
}


func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}


func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}


func Clamp(value, min, max float64) float64 {
	return math.Max(min, math.Min(max, value))
}
