package main

import (
	"github.com/jrevanaldi-ai/gowa-bot/commands/debug"
	"github.com/jrevanaldi-ai/gowa-bot/commands/download"
	general "github.com/jrevanaldi-ai/gowa-bot/commands/general"
	"github.com/jrevanaldi-ai/gowa-bot/commands/jadibot"
	"github.com/jrevanaldi-ai/gowa-bot/commands/maker"
	"github.com/jrevanaldi-ai/gowa-bot/commands/owner"
	"github.com/jrevanaldi-ai/gowa-bot/commands/utility"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

func registerCommands(registry *lib.CommandRegistry) {

	registry.Register(utility.PingMetadata, utility.PingHandler)

	registry.Register(general.MenuMetadata, general.MenuHandler)

	registry.Register(general.HelpMetadata, general.HelpHandler)

	registry.Register(general.GetppMetadata, general.GetppHandler)

	registry.Register(general.DonasiMetadata, general.DonasiHandler)
	registry.Register(general.CekDonasiMetadata, general.CekDonasiHandler)

	registry.Register(general.LuneMetadata, general.LuneHandler)

	registry.Register(debug.CheckEphemeralMetadata, debug.CheckEphemeralHandler)

	registry.Register(owner.ExecMetadata, owner.ExecHandler)
	registry.Register(owner.EvalMetadata, owner.EvalHandler)

	registry.Register(owner.SetmodeMetadata, owner.SetmodeHandler)

	registry.Register(owner.InfoserverMetadata, owner.InfoserverHandler)

	registry.Register(owner.ReactMetadata, owner.ReactHandler)

	registry.Register(owner.BangroupMetadata, owner.BangroupHandler)
	registry.Register(owner.UnbangroupMetadata, owner.UnbangroupHandler)
	registry.Register(owner.BanuserMetadata, owner.BanuserHandler)
	registry.Register(owner.UnbanuserMetadata, owner.UnbanuserHandler)

	registry.Register(owner.SetprefixMetadata, owner.SetprefixHandler)

	registry.Register(owner.JoinMetadata, owner.JoinHandler)

	registry.Register(jadibot.JadibotMetadata, jadibot.JadibotHandler)
	registry.Register(jadibot.ListJadibotMetadata, jadibot.ListJadibotHandler)
	registry.Register(jadibot.StopJadibotMetadata, jadibot.StopJadibotHandler)
	registry.Register(jadibot.PauseJadibotMetadata, jadibot.PauseJadibotHandler)
	registry.Register(jadibot.ResumeJadibotMetadata, jadibot.ResumeJadibotHandler)
	registry.Register(jadibot.RemoveJadibotMetadata, jadibot.RemoveJadibotHandler)

	registry.Register(download.PlayMetadata, download.PlayHandler)

	registry.Register(download.SpotifyMetadata, download.SpotifyHandler)

	registry.Register(download.InstagramMetadata, download.InstagramHandler)
	registry.Register(download.GitHubMetadata, download.GitHubHandler)


	registry.Register(download.TikTokMetadata, download.TikTokHandler)

	registry.Register(download.TTSearchMetadata, download.TTSearchHandler)

	registry.Register(maker.BratMetadata, maker.BratHandler)
}
