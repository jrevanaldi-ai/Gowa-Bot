package general

import (
	"context"
	"fmt"
	"strings"

	"go.mau.fi/whatsmeow"
	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

const DefaultWelcomeText = `🎉 Selamat datang @user!

Senang kamu bergabung di @group ✨
Kamu adalah member ke-@count di sini.

Ketik .menu untuk lihat daftar command bot.
Selamat menikmati! 🚀`

const DefaultGoodbyeText = `👋 Sampai jumpa @user...

Terima kasih sudah pernah jadi bagian dari @group.
Sekarang tersisa @count member.

Semoga ketemu lagi di lain kesempatan 💫`

var WelcomeMetadata = &lib.CommandMetadata{
	Cmd:       "welcome",
	Tag:       "group",
	Desc:      "Atur pesan selamat datang group. Placeholder: @user @group @count @desc",
	Example:   ".welcome on  /  .welcome set <text>",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"wc"},
}

func WelcomeHandler(ctx *lib.CommandContext) error {
	return handleWelcomeGoodbye(ctx, "welcome")
}

var GoodbyeMetadata = &lib.CommandMetadata{
	Cmd:       "goodbye",
	Tag:       "group",
	Desc:      "Atur pesan perpisahan group. Placeholder: @user @group @count @desc",
	Example:   ".goodbye on  /  .goodbye set <text>",
	Hidden:    false,
	OwnerOnly: false,
	Alias:     []string{"gb", "leave"},
}

func GoodbyeHandler(ctx *lib.CommandContext) error {
	return handleWelcomeGoodbye(ctx, "goodbye")
}

func handleWelcomeGoodbye(ctx *lib.CommandContext, kind string) error {
	reply := func(text string) error {
		_, err := ctx.SendMessage(helper.CreateSimpleReply(text, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	if !ctx.IsGroup {
		return reply("Command ini hanya bisa digunakan di group.")
	}

	dbManager, ok := ctx.BotClient.GetDBManager().(*helper.DatabaseManager)
	if !ok || dbManager == nil {
		return fmt.Errorf("database manager tidak tersedia")
	}

	if !ctx.IsOwner {
		isAdmin, err := isGroupAdmin(ctx.Client, ctx.Chat, ctx.Sender)
		if err != nil {
			return reply(fmt.Sprintf("Gagal cek admin: %v", err))
		}
		if !isAdmin {
			return reply("Hanya admin group atau owner bot yang bisa pakai command ini.")
		}
	}

	sub := ""
	if len(ctx.Args) > 0 {
		sub = strings.ToLower(ctx.Args[0])
	}
	groupJID := ctx.Chat.String()

	switch sub {
	case "on", "enable":
		cfg, err := dbManager.GetWelcomeConfig(groupJID)
		if err != nil {
			return fmt.Errorf("failed to read config: %w", err)
		}

		usedDefault := false
		if kind == "welcome" {
			if cfg.WelcomeText == "" {
				if err := dbManager.SetWelcomeText(groupJID, DefaultWelcomeText); err != nil {
					return fmt.Errorf("failed to seed default: %w", err)
				}
				usedDefault = true
			} else if err := dbManager.SetWelcomeEnabled(groupJID, true); err != nil {
				return fmt.Errorf("failed to enable: %w", err)
			}
		} else {
			if cfg.GoodbyeText == "" {
				if err := dbManager.SetGoodbyeText(groupJID, DefaultGoodbyeText); err != nil {
					return fmt.Errorf("failed to seed default: %w", err)
				}
				usedDefault = true
			} else if err := dbManager.SetGoodbyeEnabled(groupJID, true); err != nil {
				return fmt.Errorf("failed to enable: %w", err)
			}
		}

		if usedDefault {
			defTxt := DefaultWelcomeText
			if kind == "goodbye" {
				defTxt = DefaultGoodbyeText
			}
			return reply(fmt.Sprintf("✓ %s diaktifkan pakai template bawaan Gowa-Bot.\n\nPreview:\n%s\n\nKalau mau custom: .%s set <pesan>", strings.Title(kind), defTxt, kind))
		}
		return reply(fmt.Sprintf("✓ %s diaktifkan (pakai pesan yang sudah ada).", strings.Title(kind)))

	case "off", "disable":
		var err error
		if kind == "welcome" {
			err = dbManager.SetWelcomeEnabled(groupJID, false)
		} else {
			err = dbManager.SetGoodbyeEnabled(groupJID, false)
		}
		if err != nil {
			return fmt.Errorf("failed to disable: %w", err)
		}
		return reply(fmt.Sprintf("✓ %s dinonaktifkan.", strings.Title(kind)))

	case "set":
		if len(ctx.Args) < 2 {
			defTxt := DefaultWelcomeText
			if kind == "goodbye" {
				defTxt = DefaultGoodbyeText
			}
			return reply(fmt.Sprintf("Format: .%s set <pesan>\n\nPlaceholder:\n@user — mention user\n@group — nama group\n@count — total member\n@desc — deskripsi group\n\nTemplate bawaan (pakai .%s on aja kalau mau ini):\n%s", kind, kind, defTxt))
		}
		text := strings.TrimSpace(strings.TrimPrefix(ctx.Message, ctx.Args[0]))
		text = strings.TrimSpace(strings.TrimPrefix(text, ctx.Args[0]))
		// safer: rebuild from raw args
		text = strings.Join(ctx.Args[1:], " ")
		if text == "" {
			return reply("Pesan tidak boleh kosong.")
		}
		var err error
		if kind == "welcome" {
			err = dbManager.SetWelcomeText(groupJID, text)
		} else {
			err = dbManager.SetGoodbyeText(groupJID, text)
		}
		if err != nil {
			return fmt.Errorf("failed to set: %w", err)
		}
		return reply(fmt.Sprintf("✓ Pesan %s disimpan & diaktifkan.\n\nPreview:\n%s", kind, text))

	case "show", "info", "":
		cfg, err := dbManager.GetWelcomeConfig(groupJID)
		if err != nil {
			return fmt.Errorf("failed to get config: %w", err)
		}
		var enabled bool
		var text string
		if kind == "welcome" {
			enabled = cfg.WelcomeEnabled
			text = cfg.WelcomeText
		} else {
			enabled = cfg.GoodbyeEnabled
			text = cfg.GoodbyeText
		}
		status := "OFF"
		if enabled {
			status = "ON"
		}
		if text == "" {
			text = "(belum diset)"
		}
		out := fmt.Sprintf("%s — %s\n\nPesan:\n%s\n\nCommand:\n.%s on / off\n.%s set <pesan>",
			strings.ToUpper(kind), status, text, kind, kind)
		return reply(out)

	default:
		return reply(fmt.Sprintf("Subcommand tidak dikenal: %s\n\nPakai: .%s on|off|set|show", sub, kind))
	}
}

func isGroupAdmin(cli *whatsmeow.Client, groupJID, userJID interface{ String() string }) (bool, error) {
	if cli == nil {
		return false, fmt.Errorf("client nil")
	}
	gJID := lib.StringToJID(groupJID.String())
	info, err := cli.GetGroupInfo(context.Background(), gJID)
	if err != nil {
		return false, err
	}
	userStr := userJID.String()
	userUser := strings.Split(userStr, "@")[0]
	for _, p := range info.Participants {
		if !p.IsAdmin && !p.IsSuperAdmin {
			continue
		}
		if p.JID.String() == userStr || p.JID.User == userUser {
			return true, nil
		}
		if !p.PhoneNumber.IsEmpty() && p.PhoneNumber.User == userUser {
			return true, nil
		}
		if !p.LID.IsEmpty() && p.LID.User == userUser {
			return true, nil
		}
	}
	return false, nil
}
