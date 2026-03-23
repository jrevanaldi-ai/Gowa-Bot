package lib

import (
	"context"

	"github.com/jrevanaldi-ai/gowa"
	"github.com/jrevanaldi-ai/gowa/proto/waE2E"
	"github.com/jrevanaldi-ai/gowa/types"
)

// CommandMetadata adalah metadata untuk setiap fitur/command
type CommandMetadata struct {
	Cmd       string   // Nama command utama
	Tag       string   // Kategori/tag command
	Desc      string   // Deskripsi command
	Example   string   // Contoh penggunaan
	Hidden    bool     // Apakah command hidden (tidak tampil di menu)
	OwnerOnly bool     // Apakah hanya owner yang bisa menggunakan
	Alias     []string // Alias command (hardcoded, hanya 1 yang ditampilkan di menu)
}

// CommandContext adalah context yang diteruskan ke setiap command handler
type CommandContext struct {
	Ctx              context.Context
	Client           *gowa.Client
	Sender           types.JID
	Chat             types.JID
	PushName         string
	IsGroup          bool
	IsOwner          bool
	Message          string
	Args             []string
	MessageID        types.MessageID // ID pesan yang direply
	EphemeralWrapper func(ctx context.Context, jid types.JID, msg *waE2E.Message) (*waE2E.Message, error)
}

// SendMessage mengirim pesan dengan ephemeral wrapper jika tersedia
func (c *CommandContext) SendMessage(message *waE2E.Message) (interface{}, error) {
	// Gunakan ephemeral wrapper jika tersedia
	if c.EphemeralWrapper != nil && c.IsGroup {
		wrappedMsg, err := c.EphemeralWrapper(c.Ctx, c.Chat, message)
		if err != nil {
			// Log error tapi lanjutkan dengan pesan asli
			wrappedMsg = message
		}
		message = wrappedMsg
	}

	return c.Client.SendMessage(c.Ctx, c.Chat, message)
}

// SendResponse adalah response dari SendMessage
type SendResponse struct {
	ID        string
	Timestamp interface{}
}

// CommandHandler adalah tipe function untuk handler command
type CommandHandler func(ctx *CommandContext) error

// CommandRegistry menyimpan semua command yang terdaftar
type CommandRegistry struct {
	commands map[string]*CommandMetadata
	handlers map[string]CommandHandler
}

// NewCommandRegistry membuat registry command baru
func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		commands: make(map[string]*CommandMetadata),
		handlers: make(map[string]CommandHandler),
	}
}

// Register mendaftarkan command baru ke registry
func (r *CommandRegistry) Register(metadata *CommandMetadata, handler CommandHandler) {
	// Register main command
	r.commands[metadata.Cmd] = metadata
	r.handlers[metadata.Cmd] = handler

	// Register aliases
	for _, alias := range metadata.Alias {
		r.commands[alias] = metadata
		r.handlers[alias] = handler
	}
}

// GetCommand mendapatkan metadata command
func (r *CommandRegistry) GetCommand(cmd string) (*CommandMetadata, bool) {
	meta, ok := r.commands[cmd]
	return meta, ok
}

// GetHandler mendapatkan handler command
func (r *CommandRegistry) GetHandler(cmd string) (CommandHandler, bool) {
	handler, ok := r.handlers[cmd]
	return handler, ok
}

// GetAllCommands mendapatkan semua command yang tidak hidden
func (r *CommandRegistry) GetAllCommands() []*CommandMetadata {
	var commands []*CommandMetadata
	seen := make(map[string]bool)

	for _, meta := range r.commands {
		if !meta.Hidden && !seen[meta.Cmd] {
			seen[meta.Cmd] = true
			commands = append(commands, meta)
		}
	}

	return commands
}

// GetCommandsByTag mendapatkan command berdasarkan tag
func (r *CommandRegistry) GetCommandsByTag(tag string) []*CommandMetadata {
	var commands []*CommandMetadata
	seen := make(map[string]bool)

	for _, meta := range r.commands {
		if !meta.Hidden && meta.Tag == tag && !seen[meta.Cmd] {
			seen[meta.Cmd] = true
			commands = append(commands, meta)
		}
	}

	return commands
}
