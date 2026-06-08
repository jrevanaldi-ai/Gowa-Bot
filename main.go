package main

import (
	"bufio"
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"

	"github.com/jrevanaldi-ai/gowa-bot/client"
	general "github.com/jrevanaldi-ai/gowa-bot/commands/general"
	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
	"github.com/jrevanaldi-ai/gowa-bot/webserver"
)

var (
	logLevel         = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	pairCode         = flag.String("pair", "", "Pairing code (8 karakter)")
	phone            = flag.String("phone", "", "Nomor telepon untuk pairing (format: 62xxx)")
	dbPath           = flag.String("db", "gowa-bot.db", "Path ke database file")
	selfMode         = flag.Bool("self", false, "Self mode - bot merespon pesan dari diri sendiri")
	mustikaPayAPIKey = flag.String("mustika-api-key", "", "MustikaPay API Key untuk pembayaran")
	aiAPIKey         = flag.String("ai-api-key", "", "API Key untuk AI (Claude)")
	webAddr          = flag.String("web-addr", ":8080", "HTTP dashboard bind address (e.g. :8080 atau 127.0.0.1:8080)")
)

func main() {
	flag.Parse()

	loadEnvFile()

	helper.Banner()

	logger := helper.NewLogger("Main")
	logger.Info("Starting Gowa-Bot...")

	registry := lib.NewCommandRegistry()

	registerCommands(registry)

	dbManager, err := helper.NewDatabaseManager(*dbPath)
	if err != nil {
		logger.Error("Failed to create database manager: %v", err)
		return
	}
	defer dbManager.Close()

	if *mustikaPayAPIKey == "" {

		*mustikaPayAPIKey = os.Getenv("GOWA_BOT_MUSTIKA_API_KEY")
	}

	if *mustikaPayAPIKey != "" {
		general.SetMustikaPayAPIKey(*mustikaPayAPIKey)
		logger.Info("MustikaPay payment integration enabled")
	}

	if *aiAPIKey == "" {
		*aiAPIKey = os.Getenv("GOWA_BOT_AI_API_KEY")
	}

	if *aiAPIKey != "" {
		aiSvc := helper.NewAIService(*aiAPIKey, "", "", "", registry, nil, nil, nil, nil)
		general.SetAIService(aiSvc)
		logger.Info("AI service (Claude) enabled")
	}

	gowaLog := &formatLogger{logger: logger}

	clientFactory := func(registry *lib.CommandRegistry, owners []string, gowaClient *whatsmeow.Client) lib.BotClientInterface {
		botClient := client.NewBotClient(registry, &client.BotConfig{
			Owners:      []string{},
			Prefixes:    []string{"."},
			MaxWorkers:  10,
			EnableCache: true,
			SelfMode:    false,
			IsMainBot:   false,
			DBManager:   dbManager,
		})
		botClient.SetClient(gowaClient)
		return botClient
	}

	jadibotSessionManager := helper.NewJadibotSessionManager(dbManager, registry, gowaLog, logger, clientFactory)

	logger.Info("Resuming active jadibots...")
	activeJadibots, err := dbManager.GetActiveJadibot()
	if err != nil {
		logger.Warning("Failed to get active jadibots: %v", err)
	} else {
		for _, jadibot := range activeJadibots {
			logger.Info("Resuming jadibot: %s", jadibot.ID)
			_, err := jadibotSessionManager.StartJadibot(context.Background(), jadibot.ID, jadibot.PhoneNumber)
			if err != nil {
				logger.Warning("Failed to resume jadibot %s: %v", jadibot.ID, err)
			}
		}
	}

	botClient := client.NewBotClient(registry, &client.BotConfig{
		Owners:                getOwnerNumbers(),
		Prefixes:              []string{"."},
		MaxWorkers:            10,
		EnableCache:           true,
		SelfMode:              *selfMode,
		IsMainBot:             true,
		JadibotSessionManager: jadibotSessionManager,
		DBManager:             dbManager,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cli := connectWhatsApp(ctx, logger, botClient)
	if cli == nil {
		logger.Error("Failed to connect to WhatsApp")
		return
	}

	botClient.SetClient(cli)

	webSrv := webserver.New(*webAddr)
	webSrv.Registry = registry
	webSrv.Client = cli
	webSrv.DBManager = dbManager
	webSrv.JadibotMgr = jadibotSessionManager
	webSrv.Activity = botClient.Activity
	webSrv.GetSelfMode = botClient.GetSelfMode
	webSrv.GetPrefixes = botClient.GetPrefixes

	go func() {
		logger.Success("Web dashboard at http://localhost%s", *webAddr)
		if err := webSrv.Start(ctx); err != nil {
			logger.Warning("Web server stopped: %v", err)
		}
	}()

	logger.Success("Gowa-Bot is ready!")

	logger.Info("Press Ctrl+C to stop")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down...")

	logger.Info("Stopping all jadibots...")
	jadibotSessionManager.StopAll()

	cancel()
	cli.Disconnect()

	time.Sleep(2 * time.Second)
	logger.Success("Gowa-Bot stopped")
}

func getOwnerNumbers() []string {
	owners := os.Getenv("GOWA_BOT_OWNERS")
	if owners == "" {
		logger := helper.NewLogger("Main")
		logger.Warning("GOWA_BOT_OWNERS not set in .env or flags. Bot will have no owners configured.")
		return []string{}
	}

	result := make([]string, 0)
	for _, owner := range strings.Split(owners, ",") {
		owner = strings.TrimSpace(owner)
		if owner != "" {
			result = append(result, owner)
		}
	}
	return result
}

func connectWhatsApp(ctx context.Context, logger *helper.Logger, botClient *client.BotClient) *whatsmeow.Client {

	gowaLog := &formatLogger{logger: logger}

	container, err := sqlstore.New(ctx, "sqlite3", *dbPath+"?_foreign_keys=on", gowaLog)
	if err != nil {
		logger.Error("Failed to create store: %v", err)
		return nil
	}

	device, err := container.GetFirstDevice(ctx)
	if err != nil {
		logger.Error("Failed to get device: %v", err)
		return nil
	}

	if device == nil {
		device = container.NewDevice()
	}

	cli := whatsmeow.NewClient(device, gowaLog)

	cli.AddEventHandler(botClient.EventHandler)

	if err := cli.Connect(); err != nil {
		logger.Error("Failed to connect: %v", err)
		return nil
	}

	if cli.Store.ID != nil {
		logger.Success("Already paired as %s", cli.Store.ID.String())
		return cli
	}

	if *phone == "" {
		logger.Error("Phone number is required for pairing. Use -phone flag")
		return nil
	}

	time.Sleep(1 * time.Second)

	if *pairCode != "" {
		logger.Warning("Custom pair code (-pair) not supported by upstream whatsmeow, ignoring.")
	}
	code, err := cli.PairPhone(ctx, *phone, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
	if err != nil {
		logger.Error("Failed to pair: %v", err)
		return nil
	}

	logger.Info("Pairing code: %s", code)
	logger.Info("Enter this code in your WhatsApp app (Linked Devices)")

	maxWait := 160 * time.Second
	startTime := time.Now()

	for time.Since(startTime) < maxWait {
		if cli.Store.ID != nil {
			logger.Success("Successfully paired as %s", cli.Store.ID.String())
			return cli
		}
		time.Sleep(1 * time.Second)
	}

	logger.Error("Pairing timeout")
	return nil
}

func loadEnvFile() {
	file, err := os.Open(".env")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		line = strings.TrimPrefix(line, "export ")

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, "\"")
		value = strings.Trim(value, "'")

		os.Setenv(key, value)
	}
}

type formatLogger struct {
	logger *helper.Logger
}

func (l *formatLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debug(format, args...)
}

func (l *formatLogger) Infof(format string, args ...interface{}) {
	l.logger.Info(format, args...)
}

func (l *formatLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warning(format, args...)
}

func (l *formatLogger) Errorf(format string, args ...interface{}) {
	l.logger.Error(format, args...)
}

func (l *formatLogger) Sub(module string) waLog.Logger {
	return &formatLogger{logger: helper.NewLogger(module)}
}
