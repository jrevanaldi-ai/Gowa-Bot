package owner

import (
	"os/exec"
	"strings"

	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)


var ExecMetadata = &lib.CommandMetadata{
	Cmd:       "exec",
	Tag:       "owner",
	Desc:      "Exec command (owner only)",
	Example:   "$ls -la",
	Hidden:    true,
	OwnerOnly: true,
	Alias:     []string{"$"},
}


func ExecHandler(ctx *lib.CommandContext) error {

	if !ctx.IsOwner {
		return nil
	}


	if len(ctx.Args) == 0 {
		message := "Usage: $<command>\nExample: $ls -la"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	cmdStr := strings.Join(ctx.Args, " ")


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


	message := string(output)
	if message == "" {
		message = "✓ Command executed successfully (no output)"
	}
	_, err = ctx.SendMessage(helper.CreateSimpleReply("✓ Output:\n```\n"+message+"```", ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return err
}


func IsExecCommand(msg string) bool {
	return strings.HasPrefix(msg, "$") && len(msg) > 1
}


func ParseExecCommand(msg string) []string {
	if !IsExecCommand(msg) {
		return nil
	}

	msg = strings.TrimPrefix(msg, "$")
	return strings.Fields(msg)
}
