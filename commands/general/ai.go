package general

import (
	"fmt"
	"reflect"
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
		message := "AI tidak tersedia.\n\n" +
			"- AI service belum dikonfigurasi\n" +
			"- Hubungi owner untuk mengaktifkan"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	if len(ctx.Args) == 0 {
		message := "Masukkan pertanyaan.\n\n" +
			"Usage:\n" +
			"- .lune <pertanyaan> - Tanya AI\n" +
			"- .lune reset - Reset percakapan\n\n" +
			"Contoh:\n" +
			"- .lune apa itu golang?\n" +
			"- .lune buatkan pantun lucu"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	query := strings.Join(ctx.Args, " ")

	if strings.ToLower(query) == "reset" {
		cache := ctx.BotClient.GetCache().(*helper.Cache)
		cache.Delete("ai_history_" + ctx.Sender.String())
		_, err := ctx.SendMessage(helper.CreateSimpleReply("Berhasil me-reset percakapan.", ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	cache := ctx.BotClient.GetCache().(*helper.Cache)
	historyKey := "ai_history_" + ctx.Sender.String()

	var history []helper.AIMessage
	if val, found := cache.Get(historyKey); found {
		history = val.([]helper.AIMessage)
	}

	history = append(history, helper.AIMessage{
		Role:    "user",
		Content: query,
	})

	maxHistory := 15
	if len(history) > maxHistory {
		history = history[len(history)-maxHistory:]
	}

	for len(history) > 0 && history[0].Role != "user" {
		history = history[1:]
	}

	result, err := aiService.EnhancedChatWithHistory(history, ctx)
	if err != nil {
		errorMsg := "Gagal mendapatkan respon AI.\n\n" +
			fmt.Sprintf("Error: %s", err.Error())
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	history = append(history, helper.AIMessage{
		Role:    "assistant",
		Content: result,
	})

	if len(history) > maxHistory {
		history = history[len(history)-maxHistory:]
	}
	cache.Set(historyKey, history, 30*time.Minute)

	sentResp, err := ctx.SendMessage(helper.CreateSimpleReply(result, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	if err != nil {
		return fmt.Errorf("failed to send AI response: %w", err)
	}

	if sentMsgID := extractSentMessageID(sentResp); sentMsgID != "" {
		helper.RegisterAIMessage(cache, ctx.Chat.String(), sentMsgID)
	}

	return nil
}

func extractSentMessageID(resp interface{}) string {
	if resp == nil {
		return ""
	}
	v := reflect.ValueOf(resp)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return ""
	}
	idField := v.FieldByName("ID")
	if !idField.IsValid() || idField.Kind() != reflect.String {
		return ""
	}
	return idField.String()
}
