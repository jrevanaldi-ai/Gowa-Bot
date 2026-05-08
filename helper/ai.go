package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/jrevanaldi-ai/gowa"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

type AIService struct {
	BaseURL      string
	APIKey       string
	Model        string
	SystemPrompt string
	Client       *http.Client
	mu           sync.RWMutex

	registry   *lib.CommandRegistry
	dispatcher *lib.Dispatcher
	client     *gowa.Client
	jadibotMgr lib.JadibotSessionManagerInterface
	dbManager  interface{}
}

type AIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AIRequest struct {
	Model     string      `json:"model"`
	MaxTokens int         `json:"max_tokens"`
	System    string      `json:"system"`
	Messages  []AIMessage `json:"messages"`
}

type AIContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type AIResponse struct {
	ID           string      `json:"id"`
	Type         string      `json:"type"`
	Role         string      `json:"role"`
	Model        string      `json:"model"`
	Content      []AIContent `json:"content"`
	StopReason   string      `json:"stop_reason"`
	StopSequence string      `json:"stop_sequence"`
	Usage        AIUsage     `json:"usage"`
}

type AIUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type AIErrorResponse struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

func NewAIService(apiKey, baseURL, model, systemPrompt string, registry *lib.CommandRegistry, dispatcher *lib.Dispatcher, client *gowa.Client, jadibotMgr lib.JadibotSessionManagerInterface, dbManager interface{}) *AIService {
	if baseURL == "" {
		baseURL = "https://api.yardansh.com/api/anthropic/v1/messages"
	}
	if model == "" {
		model = "claude-opus-4.7"
	}
	if systemPrompt == "" {
		systemPrompt = "Berikan jawaban langsung, singkat, dan jelas.\n\n" +
			"ATURAN KERJA:\n" +
			"1. Selalu jawab dalam bahasa Indonesia.\n" +
			"2. JANGAN MENGGUNAKAN KARAKTER TAMBAHAN UNTUK FORMATTING. Dilarang menggunakan bintang (**), garis bawah (_), atau backticks (```). Hanya gunakan huruf, angka, tanda koma (,), dan tanda titik (.).\n" +
			"3. JANGAN memberikan preamble seperti 'Tentu, ini jawabannya' atau 'Halo'. Langsung ke inti jawaban.\n" +
			"4. Jika diminta memutar musik/video atau konten media, WAJIB menghasilkan command: `.nama_command [argumen]` atau `$nama_command [argumen]` (untuk owner).\n" +
			"5. JANGAN hanya memberi info link, eksekusi melalui command.\n\n" +
			"DAFTAR COMMAND:\n" +
			"Utility: .ping\n" +
			"General: .menu, .help, .getpp, .donasi, .cekdonasi, .lune\n" +
			"Owner: $, .checkephemeral, .setmode, .infoserver, .react, .bangroup, .unbangroup, .banuser, .unbanuser, .setprefix\n" +
			"Download: .play, .spotify, .instagram, .tiktok, .ttsearch\n" +
			"Jadibot: .jadibot, .listjadibot, .stopjadibot, .pausejadibot, .resumejadibot, .removejadibot"
	}

	return &AIService{
		BaseURL:      baseURL,
		APIKey:       apiKey,
		Model:        model,
		SystemPrompt: systemPrompt,
		Client: &http.Client{
			Timeout: 120 * time.Second,
		},
		registry:   registry,
		dispatcher: dispatcher,
		client:     client,
		jadibotMgr: jadibotMgr,
		dbManager:  dbManager,
	}
}

func (ai *AIService) SetSystemPrompt(prompt string) {
	ai.mu.Lock()
	defer ai.mu.Unlock()
	ai.SystemPrompt = prompt
}

func (ai *AIService) GetSystemPrompt() string {
	ai.mu.RLock()
	defer ai.mu.RUnlock()
	return ai.SystemPrompt
}

func (ai *AIService) Chat(messages []AIMessage) (string, error) {
	ai.mu.RLock()
	systemPrompt := ai.SystemPrompt
	model := ai.Model
	baseURL := ai.BaseURL
	apiKey := ai.APIKey
	ai.mu.RUnlock()

	if apiKey == "" {
		return "", fmt.Errorf("API key not configured")
	}

	req := AIRequest{
		Model:     model,
		MaxTokens: 2000,
		System:    systemPrompt,
		Messages:  messages,
	}

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := ai.Client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp AIErrorResponse
		if jsonErr := json.Unmarshal(body, &errResp); jsonErr == nil {
			return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, errResp.Error.Message)
		}
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var aiResp AIResponse
	if err := json.Unmarshal(body, &aiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(aiResp.Content) == 0 {
		return "", fmt.Errorf("empty response from AI")
	}

	var result string
	for _, content := range aiResp.Content {
		if content.Type == "text" {
			result += content.Text
		}
	}

	return result, nil
}

func (ai *AIService) EnhancedChat(userMessage string, ctx *lib.CommandContext) (string, error) {
	return ai.EnhancedChatWithHistory([]AIMessage{{Role: "user", Content: userMessage}}, ctx)
}

func (ai *AIService) EnhancedChatWithHistory(messages []AIMessage, ctx *lib.CommandContext) (string, error) {

	aiResponse, err := ai.Chat(messages)
	if err != nil {
		return "", err
	}

	executedCommands, err := ai.detectAndExecuteCommands(aiResponse, ctx)
	if err != nil {

		return aiResponse + "\n\n⚠️ *Catatan:* Terjadi error saat mengeksekusi command yang terdeteksi.", nil
	}

	if executedCommands != "" {
		return aiResponse + "\n\n" + executedCommands, nil
	}

	return aiResponse, nil
}

func (ai *AIService) detectAndExecuteCommands(text string, ctx *lib.CommandContext) (string, error) {

	commandRegex := regexp.MustCompile(`(?:\.|\\$)([a-zA-Z]+)(?:\s+([^\n]*))?`)
	matches := commandRegex.FindAllStringSubmatch(text, -1)

	if len(matches) == 0 {
		return "", nil
	}

	var executed []string
	var errors []string

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		cmd := match[1]
		argsStr := ""
		if len(match) >= 3 && match[2] != "" {
			argsStr = match[2]
		}

		var args []string
		if argsStr != "" {
			args = strings.Fields(argsStr)
		}

		result, err := ai.executeCommand(cmd, args, ctx)
		if err != nil {
			errors = append(errors, fmt.Sprintf("❌ Gagal menjalankan %s: %v", cmd, err))
		} else if result != "" {
			executed = append(executed, fmt.Sprintf("✅ %s: %s", cmd, result))
		}
	}

	var resultParts []string
	if len(executed) > 0 {
		resultParts = append(resultParts, executed...)
	}
	if len(errors) > 0 {
		resultParts = append(resultParts, errors...)
	}

	if len(resultParts) == 0 {
		return "", nil
	}

	return strings.Join(resultParts, "\n"), nil
}

func (ai *AIService) executeCommand(cmdName string, args []string, ctx *lib.CommandContext) (string, error) {

	metadata, exists := ai.registry.GetCommand(cmdName)
	if !exists {

		found := false
		for _, meta := range ai.registry.GetAllCommands() {
			for _, alias := range meta.Alias {
				if alias == cmdName {
					metadata = meta
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		if !found {
			return "", fmt.Errorf("command '%s' tidak ditemukan", cmdName)
		}
	}

	handler, exists := ai.registry.GetHandler(metadata.Cmd)
	if !exists {
		return "", fmt.Errorf("handler untuk command '%s' tidak ditemukan", metadata.Cmd)
	}

	newCtx := &lib.CommandContext{
		Ctx:                   ctx.Ctx,
		Client:                ctx.Client,
		BotClient:             ctx.BotClient,
		JadibotSessionManager: ctx.JadibotSessionManager,
		Sender:                ctx.Sender,
		Chat:                  ctx.Chat,
		PushName:              ctx.PushName,
		IsGroup:               ctx.IsGroup,
		IsOwner:               ctx.IsOwner,
		Message:               "." + cmdName + " " + strings.Join(args, " "),
		Args:                  args,
		MessageID:             ctx.MessageID,
		EphemeralWrapper:      ctx.EphemeralWrapper,
		ReplyMessage:          ctx.ReplyMessage,
		Mentions:              ctx.Mentions,
	}

	err := handler(newCtx)
	if err != nil {
		return "", err
	}

	return "Command berhasil dieksekusi", nil
}

func (ai *AIService) SimpleChat(userMessage string) (string, error) {
	messages := []AIMessage{
		{Role: "user", Content: userMessage},
	}
	return ai.Chat(messages)
}
