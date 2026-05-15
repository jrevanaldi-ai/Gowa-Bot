package owner

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
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
		message := "Masukkan kode Go.\n\n" +
			"Usage:\n" +
			"- >> <kode>\n\n" +
			"Variables tersedia:\n" +
			"- ctx: *lib.CommandContext\n" +
			"- c: *gowa.Client\n" +
			"- db: *helper.DatabaseManager\n" +
			"- m: *events.Message (jika tersedia)"
		_, err := ctx.SendMessage(helper.CreateSimpleReply(message, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}

	code := strings.Join(ctx.Args, " ")

	i := interp.New(interp.Options{})

	if err := i.Use(stdlib.Symbols); err != nil {
		return fmt.Errorf("failed to load stdlib: %w", err)
	}

	if err := i.Use(Symbols); err != nil {
		return fmt.Errorf("failed to load bot symbols: %w", err)
	}

	mainExports := make(map[string]reflect.Value)
	mainExports["ctx"] = reflect.ValueOf(ctx)
	mainExports["c"] = reflect.ValueOf(ctx.Client)

	for _, pkgSyms := range Symbols {
		for name, val := range pkgSyms {
			mainExports[name] = val
		}
	}

	if db := ctx.BotClient.GetDBManager(); db != nil {
		mainExports["db"] = reflect.ValueOf(db)
	}

	if err := i.Use(interp.Exports{"main/main": mainExports}); err != nil {
		return fmt.Errorf("failed to inject main symbols: %w", err)
	}

	wrappedCode := fmt.Sprintf(`
package main
import (
	"fmt"
	"context"
	"time"
	"reflect"
	"google.golang.org/protobuf/proto"
)

func Run() interface{} {
	%s
}
`, code)

	_, err := i.Eval(wrappedCode)
	if err != nil {
		errorMsg := fmt.Sprintf("Eval Error:\n%s", err.Error())
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	v, err := i.Eval("main.Run()")
	if err != nil {
		errorMsg := fmt.Sprintf("Runtime Error:\n%s", err.Error())
		_, _ = ctx.SendMessage(helper.CreateSimpleReply(errorMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return nil
	}

	result := fmt.Sprintf("%v", v.Interface())
	if result == "<nil>" || result == "" {
		result = "Done (no return value)"
	}

	response := fmt.Sprintf("Eval Result:\n%s", result)
	_, err = ctx.SendMessage(helper.CreateSimpleReply(response, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return err
}
