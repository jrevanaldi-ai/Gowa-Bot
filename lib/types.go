package lib

import (
	"context"

	"github.com/jrevanaldi-ai/gowa"
	"github.com/jrevanaldi-ai/gowa/proto/waE2E"
	"github.com/jrevanaldi-ai/gowa/types"
)


type BotClientInterface interface {
	SetSelfMode(mode bool)
	GetSelfMode() bool
	EventHandler(evt any)
	SetClient(client *gowa.Client)
	GetDBManager() interface{}
	SetPrefixes(prefixes []string)
	GetPrefixes() []string
}


type JadibotSessionManagerInterface interface {
	CreateJadibot(ctx context.Context, ownerJID string, phoneNumber string) (string, error)
	StartJadibot(ctx context.Context, jadibotID string, phoneNumber string) (string, error)
	StopJadibot(jadibotID string) error
	PauseJadibot(jadibotID string) error
	ResumeJadibot(ctx context.Context, jadibotID string) error
	GetJadibotInfo(jadibotID string) (*JadibotInfo, error)
	GetJadibotByOwner(ownerJID string) ([]JadibotInfo, error)
	DeleteJadibot(jadibotID string) error
	IsRunning(jadibotID string) bool
	GetActiveBotsCount() int
}


type JadibotInfo struct {
	ID           string
	OwnerJID     string
	PhoneNumber  string
	SessionPath  string
	Status       string
	CreatedAt    interface{}
	StartedAt    interface{}
	LastActiveAt interface{}
}


type CommandMetadata struct {
	Cmd       string
	Tag       string
	Desc      string
	Example   string
	Hidden    bool
	OwnerOnly bool
	Alias     []string
}


type CommandContext struct {
	Ctx                     context.Context
	Client                  *gowa.Client
	BotClient               BotClientInterface
	JadibotSessionManager   JadibotSessionManagerInterface
	Sender                  types.JID
	Chat                    types.JID
	PushName                string
	IsGroup                 bool
	IsOwner                 bool
	Message                 string
	Args                    []string
	MessageID               types.MessageID
	EphemeralWrapper        func(ctx context.Context, jid types.JID, msg *waE2E.Message) (*waE2E.Message, error)
}


func (c *CommandContext) SendMessage(message *waE2E.Message) (interface{}, error) {

	if c.EphemeralWrapper != nil && c.IsGroup {
		wrappedMsg, err := c.EphemeralWrapper(c.Ctx, c.Chat, message)
		if err != nil {

			wrappedMsg = message
		}
		message = wrappedMsg
	}

	return c.Client.SendMessage(c.Ctx, c.Chat, message)
}


type SendResponse struct {
	ID        string
	Timestamp interface{}
}


type CommandHandler func(ctx *CommandContext) error


type CommandRegistry struct {
	commands map[string]*CommandMetadata
	handlers map[string]CommandHandler
}


func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		commands: make(map[string]*CommandMetadata),
		handlers: make(map[string]CommandHandler),
	}
}


func (r *CommandRegistry) Register(metadata *CommandMetadata, handler CommandHandler) {

	r.commands[metadata.Cmd] = metadata
	r.handlers[metadata.Cmd] = handler


	for _, alias := range metadata.Alias {
		r.commands[alias] = metadata
		r.handlers[alias] = handler
	}
}


func (r *CommandRegistry) GetCommand(cmd string) (*CommandMetadata, bool) {
	meta, ok := r.commands[cmd]
	return meta, ok
}


func (r *CommandRegistry) GetHandler(cmd string) (CommandHandler, bool) {
	handler, ok := r.handlers[cmd]
	return handler, ok
}


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
