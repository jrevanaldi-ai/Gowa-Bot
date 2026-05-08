package owner

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

var EvalMetadata = &lib.CommandMetadata{
	Cmd:       "eval",
	Tag:       "owner",
	Desc:      "Evaluate Go code (owner only)",
	Example:   ">> return fmt.Sprintf(\"Chat ID: %s\", ctx.Chat)",
	Hidden:    true,
	OwnerOnly: true,
	Alias:     []string{">>", "ev"},
}

func EvalHandler(ctx *lib.CommandContext) error {
	if !ctx.IsOwner {
		return nil
	}

	if len(ctx.Args) == 0 {
		message := "❌ *Masukkan kode Go!*\n\n" +
			"┌─⦿ *Usage*\n" +
			"│ • `>> <kode>`\n" +
			"└──────────────\n" +
			"*Variables tersedia:*\n" +
			"• `ctx`: *lib.CommandContext\n" +
			"• `c`: *gowa.Client\n" +
			"• `db`: *helper.DatabaseManager\n" +
			"• `m`: *events.Message (jika tersedia)"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	code := strings.Join(ctx.Args, " ")

	// Setup interpreter
	i := interp.New(interp.Options{})

	// Use standard library
	if err := i.Use(stdlib.Symbols); err != nil {
		return fmt.Errorf("failed to load stdlib: %w", err)
	}

	// Use extracted symbols
	if err := i.Use(Symbols); err != nil {
		return fmt.Errorf("failed to load bot symbols: %w", err)
	}

	// Inject variables
	exports := map[string]map[string]reflect.Value{
		"main/main": {
			"ctx": reflect.ValueOf(ctx),
			"c":   reflect.ValueOf(ctx.Client),
		},
	}
	
	// Try to get database manager if available
	if db := ctx.BotClient.GetDBManager(); db != nil {
		exports["main/main"]["db"] = reflect.ValueOf(db)
	}

	if err := i.Use(exports); err != nil {
		return fmt.Errorf("failed to load variables: %w", err)
	}

	// Prepare the code
	// Wrap in a function to allow 'return'
	wrappedCode := fmt.Sprintf(`
package main
import (
	"fmt"
	"context"
	"time"
	"github.com/jrevanaldi-ai/gowa/types"
	"github.com/jrevanaldi-ai/gowa/proto/waE2E"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

func Run(ctx *lib.CommandContext) interface{} {
	%s
}
`, code)

	// Execute
	_, err := i.Eval(wrappedCode)
	if err != nil {
		errorMsg := fmt.Sprintf("❌ *Eval Error:*\n```\n%s\n```", err.Error())
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	v, err := i.Eval("main.Run(ctx)")
	if err != nil {
		errorMsg := fmt.Sprintf("❌ *Runtime Error:*\n```\n%s\n```", err.Error())
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	// Format result
	result := fmt.Sprintf("%v", v.Interface())
	if result == "<nil>" || result == "" {
		result = "✓ Done (no return value)"
	}

	response := fmt.Sprintf("✅ *Eval Result:*\n```\n%s\n```", result)
	_, err = ctx.SendMessage(helper.CreateSimpleReply(response, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return err
}
