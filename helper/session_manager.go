package helper

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jrevanaldi-ai/gowa"
	"github.com/jrevanaldi-ai/gowa/store/sqlstore"
	"github.com/jrevanaldi-ai/gowa/types/events"
	waLog "github.com/jrevanaldi-ai/gowa/util/log"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)


type JadibotInstance struct {
	ID         string
	Client     *gowa.Client
	BotClient  lib.BotClientInterface
	Info       JadibotInfo
	CancelFunc context.CancelFunc
	mu         sync.RWMutex
}


type JadibotSessionManager struct {
	DBManager     *DatabaseManager
	Registry      *lib.CommandRegistry
	OwnerNumbers  []string
	ClientFactory func(registry *lib.CommandRegistry, owners []string, gowaClient *gowa.Client) lib.BotClientInterface
	ActiveBots    map[string]*JadibotInstance
	Logger        *Logger
	GowaLogger    waLog.Logger
	mu            sync.RWMutex
}


func NewJadibotSessionManager(dbManager *DatabaseManager, registry *lib.CommandRegistry, gowaLogger waLog.Logger, logger *Logger, clientFactory func(registry *lib.CommandRegistry, owners []string, gowaClient *gowa.Client) lib.BotClientInterface) *JadibotSessionManager {
	return &JadibotSessionManager{
		DBManager:     dbManager,
		Registry:      registry,
		ActiveBots:    make(map[string]*JadibotInstance),
		Logger:        logger,
		GowaLogger:    gowaLogger,
		ClientFactory: clientFactory,
	}
}


func (m *JadibotSessionManager) SetRegistry(registry *lib.CommandRegistry) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Registry = registry
}


func (m *JadibotSessionManager) SetOwnerNumbers(owners []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.OwnerNumbers = owners
}


func GenerateSessionPath() (string, string) {
	id := uuid.New().String()
	sessionPath := filepath.Join("sessions", "jadibot_"+id)
	return id, sessionPath
}


func (m *JadibotSessionManager) CreateJadibot(ctx context.Context, ownerJID string, phoneNumber string) (string, error) {

	id, sessionPath := GenerateSessionPath()


	if err := os.MkdirAll(sessionPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create session directory: %w", err)
	}


	info := JadibotInfo{
		ID:          id,
		OwnerJID:    ownerJID,
		PhoneNumber: phoneNumber,
		SessionPath: sessionPath,
		Status:      string(StatusStopped),
	}

	if err := m.DBManager.CreateJadibot(info); err != nil {
		return "", fmt.Errorf("failed to save jadibot to database: %w", err)
	}

	m.Logger.Info("Jadibot created: %s (ID: %s)", phoneNumber, id)
	return id, nil
}


func (m *JadibotSessionManager) StartJadibot(ctx context.Context, jadibotID string, phoneNumber string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()


	if _, exists := m.ActiveBots[jadibotID]; exists {
		return "", fmt.Errorf("jadibot %s is already running", jadibotID)
	}


	info, err := m.DBManager.GetJadibot(jadibotID)
	if err != nil {
		return "", fmt.Errorf("failed to get jadibot info: %w", err)
	}


	if err := os.MkdirAll(info.SessionPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create session directory: %w", err)
	}


	dbPath := filepath.Join(info.SessionPath, "jadibot.db")
	container, err := sqlstore.New(ctx, "sqlite3", dbPath+"?_foreign_keys=on", m.GowaLogger)
	if err != nil {
		return "", fmt.Errorf("failed to create store: %w", err)
	}


	device, err := container.GetFirstDevice(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get device: %w", err)
	}

	if device == nil {
		device = container.NewDevice()
	}


	cli := gowa.NewClient(device, m.GowaLogger)


	if err := cli.Connect(); err != nil {
		return "", fmt.Errorf("failed to connect: %w", err)
	}


	jadibotBotClient := m.ClientFactory(m.Registry, m.OwnerNumbers, cli)


	if cli.Store.ID != nil {

		m.Logger.Success("Jadibot %s already paired", jadibotID)


		if err := m.DBManager.UpdateJadibotStatus(jadibotID, StatusActive); err != nil {
			m.Logger.Error("Failed to update jadibot status: %v", err)
		}


		botCtx, cancel := context.WithCancel(context.Background())
		instance := &JadibotInstance{
			ID:         jadibotID,
			Client:     cli,
			BotClient:  jadibotBotClient,
			Info:       *info,
			CancelFunc: cancel,
		}
		m.ActiveBots[jadibotID] = instance


		cli.AddEventHandler(jadibotBotClient.EventHandler)


		go m.monitorJadibot(botCtx, instance)

		return "", nil
	}


	time.Sleep(1 * time.Second)
	pairingCode, err := cli.PairPhone(ctx, phoneNumber, true, gowa.PairClientChrome, "Chrome (Linux)")
	if err != nil {
		cli.Disconnect()
		return "", fmt.Errorf("failed to pair phone: %w", err)
	}


	botCtx, cancel := context.WithCancel(context.Background())
	instance := &JadibotInstance{
		ID:         jadibotID,
		Client:     cli,
		BotClient:  jadibotBotClient,
		Info:       *info,
		CancelFunc: cancel,
	}
	m.ActiveBots[jadibotID] = instance


	if err := m.DBManager.UpdateJadibotStatus(jadibotID, StatusActive); err != nil {
		m.Logger.Error("Failed to update jadibot status: %v", err)
	}


	go m.waitForPairingAndMonitor(botCtx, instance, jadibotID)

	return pairingCode, nil
}


func (m *JadibotSessionManager) waitForPairingAndMonitor(ctx context.Context, instance *JadibotInstance, jadibotID string) {
	maxWait := 160 * time.Second
	startTime := time.Now()

	for time.Since(startTime) < maxWait {
		select {
		case <-ctx.Done():
			return
		default:
			if instance.Client.Store.ID != nil {
				m.Logger.Success("Jadibot %s paired successfully", jadibotID)


				_ = m.DBManager.UpdateJadibotStatus(jadibotID, StatusActive)


				if instance.BotClient != nil {
					instance.Client.AddEventHandler(instance.BotClient.EventHandler)
					m.Logger.Info("BotClient event handler added for jadibot %s", jadibotID)
				}


				m.monitorJadibot(ctx, instance)
				return
			}
			time.Sleep(1 * time.Second)
		}
	}


	m.Logger.Error("Jadibot %s pairing timeout", jadibotID)
	m.StopJadibot(jadibotID)
}


func (m *JadibotSessionManager) monitorJadibot(ctx context.Context, instance *JadibotInstance) {

	instance.Client.AddEventHandler(func(evt any) {
		switch evt.(type) {
		case *events.Connected:
			m.Logger.Info("Jadibot %s connected", instance.ID)
		case *events.LoggedOut:
			m.Logger.Warning("Jadibot %s logged out", instance.ID)
			m.StopJadibot(instance.ID)
		case *events.Disconnected:
			m.Logger.Warning("Jadibot %s disconnected", instance.ID)

			go m.tryReconnect(instance)
		}
	})


	<-ctx.Done()
}


func (m *JadibotSessionManager) tryReconnect(instance *JadibotInstance) {
	for i := 0; i < 5; i++ {
		time.Sleep(time.Duration(i+1) * 5 * time.Second)

		if err := instance.Client.Connect(); err != nil {
			m.Logger.Error("Jadibot %s reconnect attempt %d failed: %v", instance.ID, i+1, err)
			continue
		}

		m.Logger.Success("Jadibot %s reconnected", instance.ID)
		return
	}

	m.Logger.Error("Jadibot %s failed to reconnect after 5 attempts", instance.ID)
	m.StopJadibot(instance.ID)
}


func (m *JadibotSessionManager) StopJadibot(jadibotID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.ActiveBots[jadibotID]
	if !exists {
		return fmt.Errorf("jadibot %s is not running", jadibotID)
	}


	if instance.CancelFunc != nil {
		instance.CancelFunc()
	}


	if instance.Client != nil {
		instance.Client.Disconnect()
	}


	delete(m.ActiveBots, jadibotID)


	if err := m.DBManager.UpdateJadibotStatus(jadibotID, StatusStopped); err != nil {
		m.Logger.Error("Failed to update jadibot status: %v", err)
	}

	m.Logger.Info("Jadibot %s stopped", jadibotID)
	return nil
}


func (m *JadibotSessionManager) PauseJadibot(jadibotID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.ActiveBots[jadibotID]
	if !exists {
		return fmt.Errorf("jadibot %s is not running", jadibotID)
	}


	if instance.CancelFunc != nil {
		instance.CancelFunc()
	}


	if instance.Client != nil {
		instance.Client.Disconnect()
	}


	delete(m.ActiveBots, jadibotID)


	if err := m.DBManager.UpdateJadibotStatus(jadibotID, StatusPaused); err != nil {
		m.Logger.Error("Failed to update jadibot status: %v", err)
	}

	m.Logger.Info("Jadibot %s paused", jadibotID)
	return nil
}


func (m *JadibotSessionManager) ResumeJadibot(ctx context.Context, jadibotID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()


	info, err := m.DBManager.GetJadibot(jadibotID)
	if err != nil {
		return fmt.Errorf("failed to get jadibot info: %w", err)
	}

	if info.Status != string(StatusPaused) {
		return fmt.Errorf("jadibot %s is not paused (current status: %s)", jadibotID, info.Status)
	}


	if _, exists := m.ActiveBots[jadibotID]; exists {
		return fmt.Errorf("jadibot %s is already running", jadibotID)
	}


	dbPath := filepath.Join(info.SessionPath, "jadibot.db")
	container, err := sqlstore.New(ctx, "sqlite3", dbPath+"?_foreign_keys=on", m.GowaLogger)
	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}


	device, err := container.GetFirstDevice(ctx)
	if err != nil {
		return fmt.Errorf("failed to get device: %w", err)
	}

	if device == nil {
		return fmt.Errorf("no device found in session")
	}


	cli := gowa.NewClient(device, m.GowaLogger)


	if err := cli.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}


	jadibotBotClient := m.ClientFactory(m.Registry, m.OwnerNumbers, cli)


	if err := m.DBManager.UpdateJadibotStatus(jadibotID, StatusActive); err != nil {
		m.Logger.Error("Failed to update jadibot status: %v", err)
	}


	botCtx, cancel := context.WithCancel(context.Background())
	instance := &JadibotInstance{
		ID:         jadibotID,
		Client:     cli,
		BotClient:  jadibotBotClient,
		Info:       *info,
		CancelFunc: cancel,
	}
	m.ActiveBots[jadibotID] = instance


	cli.AddEventHandler(jadibotBotClient.EventHandler)


	go m.monitorJadibot(botCtx, instance)

	m.Logger.Info("Jadibot %s resumed", jadibotID)
	return nil
}


func (m *JadibotSessionManager) GetActiveJadibot() ([]JadibotInfo, error) {
	return m.DBManager.GetActiveJadibot()
}


func (m *JadibotSessionManager) GetJadibotInfo(jadibotID string) (*JadibotInfo, error) {
	return m.DBManager.GetJadibot(jadibotID)
}


func (m *JadibotSessionManager) GetJadibotByOwner(ownerJID string) ([]JadibotInfo, error) {
	return m.DBManager.GetJadibotByOwner(ownerJID)
}


func (m *JadibotSessionManager) DeleteJadibot(jadibotID string) error {

	m.StopJadibot(jadibotID)


	info, err := m.DBManager.GetJadibot(jadibotID)
	if err == nil {
		if err := os.RemoveAll(info.SessionPath); err != nil {
			m.Logger.Error("Failed to remove session directory: %v", err)
		}
	}


	return m.DBManager.DeleteJadibot(jadibotID)
}


func (m *JadibotSessionManager) StopAll() {
	m.mu.RLock()
	jadibotIDs := make([]string, 0, len(m.ActiveBots))
	for id := range m.ActiveBots {
		jadibotIDs = append(jadibotIDs, id)
	}
	m.mu.RUnlock()

	for _, id := range jadibotIDs {
		m.StopJadibot(id)
	}
}


func (m *JadibotSessionManager) GetActiveBotsCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.ActiveBots)
}


func (m *JadibotSessionManager) IsRunning(jadibotID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.ActiveBots[jadibotID]
	return exists
}
