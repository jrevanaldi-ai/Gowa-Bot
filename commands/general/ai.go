package general

import (
	"fmt"
	"strings"
	"time"

	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

var aiService *helper.AIService

func SetAIService(service *helper.AIService) {
	aiService = service
}

var LuneMetadata = &lib.CommandMetadata{
	Cmd:       "lune",
	Tag:       "utility",
	Desc:      "Tanya jawab dengan AI (Claude)",
	Example:   ".lune apa itu golang?",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{},
}

func LuneHandler(ctx *lib.CommandContext) error {
	if aiService == nil {
		message := "❌ *AI tidak tersedia!*\n\n" +
			"┌─⦿ *Info*\n" +
			"│ • AI service belum dikonfigurasi\n" +
			"│ • Hubungi owner untuk mengaktifkan\n" +
			"└──────────────"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	if len(ctx.Args) == 0 {
		message := "❌ *Masukkan pertanyaan!*\n\n" +
			"┌─⦿ *Usage*\n" +
			"│ • `.lune <pertanyaan>` - Tanya AI\n" +
			"│ • `.lune reset` - Reset percakapan\n" +
			"└──────────────\n\n" +
			"*📝 Contoh:*\n" +
			"• `.lune apa itu golang?`\n" +
			"• `.lune buatkan pantun lucu`"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	query := strings.Join(ctx.Args, " ")

	// Handle reset command
	if strings.ToLower(query) == "reset" {
		cache := ctx.BotClient.GetCache().(*helper.Cache)
		cache.Delete("ai_history_" + ctx.Chat.String())
		_, err := ctx.SendMessage(helper.CreateSimpleReply("✅ *Berhasil me-reset percakapan!*", ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Ambil history dari cache
	cache := ctx.BotClient.GetCache().(*helper.Cache)
	historyKey := "ai_history_" + ctx.Chat.String()
	
	var history []helper.AIMessage
	if val, found := cache.Get(historyKey); found {
		history = val.([]helper.AIMessage)
	}

	// Tambahkan pesan user ke history
	history = append(history, helper.AIMessage{
		Role:    "user",
		Content: query,
	})

	// Batasi history (misal 15 pesan terakhir)
	maxHistory := 15
	if len(history) > maxHistory {
		history = history[len(history)-maxHistory:]
	}

	// Pastikan pesan pertama adalah 'user' (Syarat Anthropic API)
	for len(history) > 0 && history[0].Role != "user" {
		history = history[1:]
	}

	// Gunakan EnhancedChatWithHistory
	result, err := aiService.EnhancedChatWithHistory(history, ctx)
	if err != nil {
		errorMsg := "❌ *Gagal mendapatkan respon AI!*\n\n" +
			"┌─⦿ *Error*\n" +
			fmt.Sprintf("│ • %s\n", err.Error()) +
			"└──────────────"
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	// Tambahkan respon AI ke history
	history = append(history, helper.AIMessage{
		Role:    "assistant",
		Content: result,
	})

	// Simpan kembali ke cache (TTL 30 menit)
	if len(history) > maxHistory {
		history = history[len(history)-maxHistory:]
	}
	cache.Set(historyKey, history, 30*time.Minute)

	_, err = ctx.SendMessage(helper.CreateSimpleReply(result, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	if err != nil {
		return fmt.Errorf("failed to send AI response: %w", err)
	}

	return nil
}
