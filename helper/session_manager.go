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

// JadibotInstance merepresentasikan instance jadibot yang sedang berjalan
type JadibotInstance struct {
	ID         string
	Client     *gowa.Client
	BotClient  lib.BotClientInterface // BotClient interface untuk handle messages
	Info       JadibotInfo
	CancelFunc context.CancelFunc
	mu         sync.RWMutex
}

// JadibotSessionManager mengelola session jadibot
type JadibotSessionManager struct {
	DBManager     *DatabaseManager
	Registry      *lib.CommandRegistry   // Command registry untuk di-share ke semua jadibot
	OwnerNumbers  []string               // Owner numbers
	ClientFactory func(registry *lib.CommandRegistry, owners []string, gowaClient *gowa.Client) lib.BotClientInterface
	ActiveBots    map[string]*JadibotInstance
	Logger        *Logger
	GowaLogger    waLog.Logger
	mu            sync.RWMutex
}

// NewJadibotSessionManager membuat session manager baru
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

// SetRegistry mengatur command registry yang akan di-share ke semua jadibot
func (m *JadibotSessionManager) SetRegistry(registry *lib.CommandRegistry) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Registry = registry
}

// SetOwnerNumbers mengatur owner numbers untuk jadibot
func (m *JadibotSessionManager) SetOwnerNumbers(owners []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.OwnerNumbers = owners
}

// GenerateSessionPath membuat path session unik untuk jadibot
func GenerateSessionPath() (string, string) {
	id := uuid.New().String()
	sessionPath := filepath.Join("sessions", "jadibot_"+id)
	return id, sessionPath
}

// CreateJadibot membuat jadibot baru dan return pairing code
func (m *JadibotSessionManager) CreateJadibot(ctx context.Context, ownerJID string, phoneNumber string) (string, error) {
	// Generate ID dan session path unik
	id, sessionPath := GenerateSessionPath()

	// Buat direktori session
	if err := os.MkdirAll(sessionPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create session directory: %w", err)
	}

	// Simpan ke database
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

// StartJadibot memulai jadibot dengan pairing code
func (m *JadibotSessionManager) StartJadibot(ctx context.Context, jadibotID string, phoneNumber string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Cek apakah jadibot sudah running
	if _, exists := m.ActiveBots[jadibotID]; exists {
		return "", fmt.Errorf("jadibot %s is already running", jadibotID)
	}

	// Get jadibot info
	info, err := m.DBManager.GetJadibot(jadibotID)
	if err != nil {
		return "", fmt.Errorf("failed to get jadibot info: %w", err)
	}

	// Buat session directory jika belum ada
	if err := os.MkdirAll(info.SessionPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create session directory: %w", err)
	}

	// Buat store container dengan session path terpisah
	dbPath := filepath.Join(info.SessionPath, "jadibot.db")
	container, err := sqlstore.New(ctx, "sqlite3", dbPath+"?_foreign_keys=on", m.GowaLogger)
	if err != nil {
		return "", fmt.Errorf("failed to create store: %w", err)
	}

	// Get or create device
	device, err := container.GetFirstDevice(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get device: %w", err)
	}

	if device == nil {
		device = container.NewDevice()
	}

	// Create client
	cli := gowa.NewClient(device, m.GowaLogger)

	// Connect
	if err := cli.Connect(); err != nil {
		return "", fmt.Errorf("failed to connect: %w", err)
	}

	// Buat BotClient untuk jadibot ini menggunakan factory
	jadibotBotClient := m.ClientFactory(m.Registry, m.OwnerNumbers, cli)

	// Cek apakah sudah paired
	if cli.Store.ID != nil {
		// Sudah paired, langsung aktif
		m.Logger.Success("Jadibot %s already paired", jadibotID)

		// Update status
		if err := m.DBManager.UpdateJadibotStatus(jadibotID, StatusActive); err != nil {
			m.Logger.Error("Failed to update jadibot status: %v", err)
		}

		// Simpan instance dengan BotClient
		botCtx, cancel := context.WithCancel(context.Background())
		instance := &JadibotInstance{
			ID:         jadibotID,
			Client:     cli,
			BotClient:  jadibotBotClient,
			Info:       *info,
			CancelFunc: cancel,
		}
		m.ActiveBots[jadibotID] = instance

		// Set event handler dengan BotClient
		cli.AddEventHandler(jadibotBotClient.EventHandler)

		// Start monitoring
		go m.monitorJadibot(botCtx, instance)

		return "", nil // Sudah paired, tidak perlu pairing code
	}

	// Belum paired, generate pairing code
	time.Sleep(1 * time.Second)
	pairingCode, err := cli.PairPhone(ctx, phoneNumber, true, gowa.PairClientChrome, "Chrome (Linux)")
	if err != nil {
		cli.Disconnect()
		return "", fmt.Errorf("failed to pair phone: %w", err)
	}

	// Simpan instance (belum aktif sampai pairing berhasil)
	botCtx, cancel := context.WithCancel(context.Background())
	instance := &JadibotInstance{
		ID:         jadibotID,
		Client:     cli,
		BotClient:  jadibotBotClient,
		Info:       *info,
		CancelFunc: cancel,
	}
	m.ActiveBots[jadibotID] = instance

	// Update status
	if err := m.DBManager.UpdateJadibotStatus(jadibotID, StatusActive); err != nil {
		m.Logger.Error("Failed to update jadibot status: %v", err)
	}

	// Start monitoring untuk tunggu pairing
	go m.waitForPairingAndMonitor(botCtx, instance, jadibotID)

	return pairingCode, nil
}

// waitForPairingAndMonitor menunggu pairing berhasil lalu monitor
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

				// Update last active
				_ = m.DBManager.UpdateJadibotStatus(jadibotID, StatusActive)

				// PENTING: Tambahkan BotClient EventHandler SETELAH pairing berhasil
				if instance.BotClient != nil {
					instance.Client.AddEventHandler(instance.BotClient.EventHandler)
					m.Logger.Info("BotClient event handler added for jadibot %s", jadibotID)
				}

				// Monitor sampai disconnected
				m.monitorJadibot(ctx, instance)
				return
			}
			time.Sleep(1 * time.Second)
		}
	}

	// Timeout
	m.Logger.Error("Jadibot %s pairing timeout", jadibotID)
	m.StopJadibot(jadibotID)
}

// monitorJadibot monitor jadibot yang aktif
func (m *JadibotSessionManager) monitorJadibot(ctx context.Context, instance *JadibotInstance) {
	// Set event handler untuk log connection
	instance.Client.AddEventHandler(func(evt any) {
		switch evt.(type) {
		case *events.Connected:
			m.Logger.Info("Jadibot %s connected", instance.ID)
		case *events.LoggedOut:
			m.Logger.Warning("Jadibot %s logged out", instance.ID)
			m.StopJadibot(instance.ID)
		case *events.Disconnected:
			m.Logger.Warning("Jadibot %s disconnected", instance.ID)
			// Coba reconnect
			go m.tryReconnect(instance)
		}
	})

	// Tunggu context done
	<-ctx.Done()
}

// tryReconnect mencoba reconnect jadibot
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

// StopJadibot menghentikan jadibot
func (m *JadibotSessionManager) StopJadibot(jadibotID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.ActiveBots[jadibotID]
	if !exists {
		return fmt.Errorf("jadibot %s is not running", jadibotID)
	}

	// Cancel context
	if instance.CancelFunc != nil {
		instance.CancelFunc()
	}

	// Disconnect client
	if instance.Client != nil {
		instance.Client.Disconnect()
	}

	// Hapus dari active bots
	delete(m.ActiveBots, jadibotID)

	// Update status di database
	if err := m.DBManager.UpdateJadibotStatus(jadibotID, StatusStopped); err != nil {
		m.Logger.Error("Failed to update jadibot status: %v", err)
	}

	m.Logger.Info("Jadibot %s stopped", jadibotID)
	return nil
}

// PauseJadibot pause jadibot (disconnect tapi tidak hapus session)
func (m *JadibotSessionManager) PauseJadibot(jadibotID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.ActiveBots[jadibotID]
	if !exists {
		return fmt.Errorf("jadibot %s is not running", jadibotID)
	}

	// Cancel context
	if instance.CancelFunc != nil {
		instance.CancelFunc()
	}

	// Disconnect client
	if instance.Client != nil {
		instance.Client.Disconnect()
	}

	// Hapus dari active bots
	delete(m.ActiveBots, jadibotID)

	// Update status di database
	if err := m.DBManager.UpdateJadibotStatus(jadibotID, StatusPaused); err != nil {
		m.Logger.Error("Failed to update jadibot status: %v", err)
	}

	m.Logger.Info("Jadibot %s paused", jadibotID)
	return nil
}

// ResumeJadibot resume jadibot yang paused
func (m *JadibotSessionManager) ResumeJadibot(ctx context.Context, jadibotID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get jadibot info
	info, err := m.DBManager.GetJadibot(jadibotID)
	if err != nil {
		return fmt.Errorf("failed to get jadibot info: %w", err)
	}

	if info.Status != string(StatusPaused) {
		return fmt.Errorf("jadibot %s is not paused (current status: %s)", jadibotID, info.Status)
	}

	// Cek apakah sudah running
	if _, exists := m.ActiveBots[jadibotID]; exists {
		return fmt.Errorf("jadibot %s is already running", jadibotID)
	}

	// Buat store container dengan session path
	dbPath := filepath.Join(info.SessionPath, "jadibot.db")
	container, err := sqlstore.New(ctx, "sqlite3", dbPath+"?_foreign_keys=on", m.GowaLogger)
	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}

	// Get device
	device, err := container.GetFirstDevice(ctx)
	if err != nil {
		return fmt.Errorf("failed to get device: %w", err)
	}

	if device == nil {
		return fmt.Errorf("no device found in session")
	}

	// Create client
	cli := gowa.NewClient(device, m.GowaLogger)

	// Connect
	if err := cli.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Buat BotClient untuk jadibot ini menggunakan factory
	jadibotBotClient := m.ClientFactory(m.Registry, m.OwnerNumbers, cli)

	// Update status
	if err := m.DBManager.UpdateJadibotStatus(jadibotID, StatusActive); err != nil {
		m.Logger.Error("Failed to update jadibot status: %v", err)
	}

	// Simpan instance
	botCtx, cancel := context.WithCancel(context.Background())
	instance := &JadibotInstance{
		ID:         jadibotID,
		Client:     cli,
		BotClient:  jadibotBotClient,
		Info:       *info,
		CancelFunc: cancel,
	}
	m.ActiveBots[jadibotID] = instance

	// Set event handler dengan BotClient
	cli.AddEventHandler(jadibotBotClient.EventHandler)

	// Start monitoring
	go m.monitorJadibot(botCtx, instance)

	m.Logger.Info("Jadibot %s resumed", jadibotID)
	return nil
}

// GetActiveJadibot mendapatkan semua jadibot yang aktif
func (m *JadibotSessionManager) GetActiveJadibot() ([]JadibotInfo, error) {
	return m.DBManager.GetActiveJadibot()
}

// GetJadibotInfo mendapatkan info jadibot
func (m *JadibotSessionManager) GetJadibotInfo(jadibotID string) (*JadibotInfo, error) {
	return m.DBManager.GetJadibot(jadibotID)
}

// GetJadibotByOwner mendapatkan jadibot berdasarkan owner
func (m *JadibotSessionManager) GetJadibotByOwner(ownerJID string) ([]JadibotInfo, error) {
	return m.DBManager.GetJadibotByOwner(ownerJID)
}

// DeleteJadibot menghapus jadibot
func (m *JadibotSessionManager) DeleteJadibot(jadibotID string) error {
	// Stop jika masih running
	m.StopJadibot(jadibotID)

	// Hapus session directory
	info, err := m.DBManager.GetJadibot(jadibotID)
	if err == nil {
		if err := os.RemoveAll(info.SessionPath); err != nil {
			m.Logger.Error("Failed to remove session directory: %v", err)
		}
	}

	// Hapus dari database
	return m.DBManager.DeleteJadibot(jadibotID)
}

// StopAll menghentikan semua jadibot
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

// GetActiveBotsCount mendapatkan jumlah jadibot yang aktif
func (m *JadibotSessionManager) GetActiveBotsCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.ActiveBots)
}

// IsRunning cek apakah jadibot sedang running
func (m *JadibotSessionManager) IsRunning(jadibotID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.ActiveBots[jadibotID]
	return exists
}
