package owner

import (
	"os/exec"
	"strings"

	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

// ExecMetadata adalah metadata untuk command exec
var ExecMetadata = &lib.CommandMetadata{
	Cmd:       "exec",
	Tag:       "owner",
	Desc:      "Exec command (owner only)",
	Example:   "$ls -la",
	Hidden:    true,
	OwnerOnly: true,
	Alias:     []string{"$"},
}

// ExecHandler menangani command exec
func ExecHandler(ctx *lib.CommandContext) error {
	// Hanya owner yang bisa menggunakan
	if !ctx.IsOwner {
		return nil
	}

	// Get command dari args
	if len(ctx.Args) == 0 {
		message := "Usage: $<command>\nExample: $ls -la"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Join semua args menjadi command
	cmdStr := strings.Join(ctx.Args, " ")

	// Exec command
	cmd := exec.Command("sh", "-c", cmdStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		message := string(output)
		if message == "" {
			message = err.Error()
		}
		_, err := ctx.SendMessage(helper.CreateSimpleReply("❌ Error:\n```\n"+message+"```", ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	// Kirim output
	message := string(output)
	if message == "" {
		message = "✓ Command executed successfully (no output)"
	}
	_, err = ctx.SendMessage(helper.CreateSimpleReply("✓ Output:\n```\n"+message+"```", ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return err
}

// IsExecCommand cek apakah pesan adalah exec command dengan prefix $
func IsExecCommand(msg string) bool {
	return strings.HasPrefix(msg, "$") && len(msg) > 1
}

// ParseExecCommand memparse exec command dari pesan
func ParseExecCommand(msg string) []string {
	if !IsExecCommand(msg) {
		return nil
	}
	// Hapus prefix $
	msg = strings.TrimPrefix(msg, "$")
	return strings.Fields(msg)
}
