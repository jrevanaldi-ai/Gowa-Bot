package lib

import (
	"context"
	"fmt"
	"sync"

	"github.com/jrevanaldi-ai/gowa"
	"github.com/jrevanaldi-ai/gowa/proto/waE2E"
	"github.com/jrevanaldi-ai/gowa/types"
)


type Dispatcher struct {
	registry  *CommandRegistry
	client    *gowa.Client
	owners    map[string]bool
	mu        sync.RWMutex
	semaphore chan struct{}
}


func NewDispatcher(registry *CommandRegistry, maxWorkers int) *Dispatcher {
	return &Dispatcher{
		registry:  registry,
		owners:    make(map[string]bool),
		semaphore: make(chan struct{}, maxWorkers),
	}
}


func (d *Dispatcher) SetClient(client *gowa.Client) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.client = client
}


func (d *Dispatcher) GetClient() *gowa.Client {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.client
}


func (d *Dispatcher) AddOwner(jid string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.owners[jid] = true
}


func (d *Dispatcher) IsOwner(jid types.JID) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.owners[jid.String()] || d.owners[jid.User]
}


func (d *Dispatcher) Dispatch(ctx context.Context, cmdCtx *CommandContext, handler CommandHandler) {

	d.semaphore <- struct{}{}
	defer func() { <-d.semaphore }()


	go func() {
		defer func() {
			if r := recover(); r != nil {
				cmdCtx.Client.Log.Errorf("Panic in command handler: %v", r)
			}
		}()

		if err := handler(cmdCtx); err != nil {
			cmdCtx.Client.Log.Errorf("Command error: %v", err)


			errorMsg := fmt.Sprintf("❌ Terjadi kesalahan: %v", err)
			_, _ = cmdCtx.Client.SendMessage(cmdCtx.Ctx, cmdCtx.Chat, &waE2E.Message{Conversation: &errorMsg})
		}
	}()
}





func InitCommands(registry *CommandRegistry) {


}


func GetCommandList(registry *CommandRegistry) map[string][]*CommandMetadata {
	commands := registry.GetAllCommands()
	commandsByTag := make(map[string][]*CommandMetadata)

	for _, cmd := range commands {
		if !cmd.Hidden && cmd.Cmd != "menu" {
			commandsByTag[cmd.Tag] = append(commandsByTag[cmd.Tag], cmd)
		}
	}

	return commandsByTag
}
