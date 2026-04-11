package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/jrevanaldi-ai/gowa"
	"github.com/jrevanaldi-ai/gowa/store/sqlstore"
	waLog "github.com/jrevanaldi-ai/gowa/util/log"

	"github.com/jrevanaldi-ai/gowa-bot/client"
	"github.com/jrevanaldi-ai/gowa-bot/commands/debug"
	"github.com/jrevanaldi-ai/gowa-bot/commands/download"
	general "github.com/jrevanaldi-ai/gowa-bot/commands/general"
	"github.com/jrevanaldi-ai/gowa-bot/commands/jadibot"
	"github.com/jrevanaldi-ai/gowa-bot/commands/owner"
	"github.com/jrevanaldi-ai/gowa-bot/commands/utility"
	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

var (
	logLevel = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	pairCode = flag.String("pair", "", "Pairing code (8 karakter)")
	phone    = flag.String("phone", "", "Nomor telepon untuk pairing (format: 62xxx)")
	dbPath   = flag.String("db", "gowa-bot.db", "Path ke database file")
	selfMode = flag.Bool("self", false, "Self mode - bot merespon pesan dari diri sendiri")
)

func main() {
	flag.Parse()


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


	gowaLog := &formatLogger{logger: logger}


	clientFactory := func(registry *lib.CommandRegistry, owners []string, gowaClient *gowa.Client) lib.BotClientInterface {
		botClient := client.NewBotClient(registry, &client.BotConfig{
			Owners:      owners,
			Prefixes:    []string{"."},
			MaxWorkers:  10,
			EnableCache: true,
			SelfMode:    false,
			DBManager:   dbManager,
		})
		botClient.SetClient(gowaClient)
		return botClient
	}

	jadibotSessionManager := helper.NewJadibotSessionManager(dbManager, registry, gowaLog, logger, clientFactory)
	jadibotSessionManager.SetOwnerNumbers(getOwnerNumbers())


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


func registerCommands(registry *lib.CommandRegistry) {

	registry.Register(utility.PingMetadata, utility.PingHandler)


	registry.Register(utility.FetchMetadata, utility.FetchHandler)


	registry.Register(utility.ThumbnailMetadata, utility.ThumbnailHandler)


	registry.Register(general.MenuMetadata, general.MenuHandler)


	registry.Register(general.HelpMetadata, general.HelpHandler)


	registry.Register(general.GetppMetadata, general.GetppHandler)


	registry.Register(debug.CheckEphemeralMetadata, debug.CheckEphemeralHandler)


	registry.Register(owner.ExecMetadata, owner.ExecHandler)


	registry.Register(owner.SetmodeMetadata, owner.SetmodeHandler)


	registry.Register(owner.InfoserverMetadata, owner.InfoserverHandler)

	registry.Register(owner.ReactMetadata, owner.ReactHandler)

	registry.Register(owner.BangroupMetadata, owner.BangroupHandler)
	registry.Register(owner.UnbangroupMetadata, owner.UnbangroupHandler)
	registry.Register(owner.BanuserMetadata, owner.BanuserHandler)
	registry.Register(owner.UnbanuserMetadata, owner.UnbanuserHandler)

	registry.Register(owner.SetprefixMetadata, owner.SetprefixHandler)

	registry.Register(jadibot.JadibotMetadata, jadibot.JadibotHandler)
	registry.Register(jadibot.ListJadibotMetadata, jadibot.ListJadibotHandler)
	registry.Register(jadibot.StopJadibotMetadata, jadibot.StopJadibotHandler)
	registry.Register(jadibot.PauseJadibotMetadata, jadibot.PauseJadibotHandler)
	registry.Register(jadibot.ResumeJadibotMetadata, jadibot.ResumeJadibotHandler)
	registry.Register(jadibot.RemoveJadibotMetadata, jadibot.RemoveJadibotHandler)


	registry.Register(download.PlayMetadata, download.PlayHandler)


	registry.Register(download.SpotifyMetadata, download.SpotifyHandler)


	registry.Register(download.InstagramMetadata, download.InstagramHandler)


	registry.Register(download.TikTokMetadata, download.TikTokHandler)


	registry.Register(download.TTSearchMetadata, download.TTSearchHandler)
}


func getOwnerNumbers() []string {
	owners := os.Getenv("GOWA_BOT_OWNERS")
	if owners == "" {

		return []string{"224983875903488"}
	}


	result := make([]string, 0)
	for _, owner := range splitString(owners, ",") {
		owner = trimSpace(owner)
		if owner != "" {
			result = append(result, owner)
		}
	}
	return result
}


func connectWhatsApp(ctx context.Context, logger *helper.Logger, botClient *client.BotClient) *gowa.Client {

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


	cli := gowa.NewClient(device, gowaLog)


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


	var code string
	if *pairCode != "" {

		code, err = cli.PairPhone(ctx, *phone, true, gowa.PairClientChrome, "Chrome (Linux)", *pairCode)
	} else {

		code, err = cli.PairPhone(ctx, *phone, true, gowa.PairClientChrome, "Chrome (Linux)")
	}

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


func splitString(s, sep string) []string {
	result := make([]string, 0)
	start := 0
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
		}
	}
	result = append(result, s[start:])
	return result
}

func trimSpace(s string) string {

	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
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
