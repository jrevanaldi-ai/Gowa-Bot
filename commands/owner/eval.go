package owner

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"reflect"
	"strings"

	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"github.com/traefik/yaegi/stdlib/unrestricted"
)

var EvalMetadata = &lib.CommandMetadata{
	Cmd:       "eval",
	Tag:       "owner",
	Desc:      "Evaluate Go code (owner only)",
	Example:   ">> ctx.Chat.String()",
	Hidden:    true,
	OwnerOnly: true,
	Alias:     []string{">>", "ev"},
}

func EvalHandler(ctx *lib.CommandContext) error {
	if !ctx.IsOwner {
		return nil
	}

	code := extractEvalCode(ctx.Message)
	if code == "" {
		return sendEvalUsage(ctx)
	}

	i := interp.New(interp.Options{})

	if err := i.Use(stdlib.Symbols); err != nil {
		return fmt.Errorf("failed to load stdlib: %w", err)
	}

	if err := i.Use(unrestricted.Symbols); err != nil {
		return fmt.Errorf("failed to load unrestricted stdlib: %w", err)
	}

	if err := i.Use(Symbols); err != nil {
		return fmt.Errorf("failed to load bot symbols: %w", err)
	}

	exports := map[string]reflect.Value{
		"Ctx":    reflect.ValueOf(ctx),
		"Client": reflect.ValueOf(ctx.Client),
	}
	hasEvent := ctx.RawEvent != nil
	if hasEvent {
		exports["Event"] = reflect.ValueOf(ctx.RawEvent)
	}
	hasRawMsg := ctx.RawMessage != nil
	if hasRawMsg {
		exports["Raw"] = reflect.ValueOf(ctx.RawMessage)
	}
	dbVal := ctx.BotClient.GetDBManager()
	hasDB := dbVal != nil
	if hasDB {
		exports["DB"] = reflect.ValueOf(dbVal)
	}
	if err := i.Use(interp.Exports{"evalctx/evalctx": exports}); err != nil {
		return fmt.Errorf("failed to inject context: %w", err)
	}

	if _, err := i.Eval(buildEvalSetup(hasDB, hasEvent, hasRawMsg)); err != nil {
		return sendEvalBlock(ctx, "Setup Error", err.Error())
	}

	if _, err := i.Eval(wrapEvalCode(code)); err != nil {
		return sendEvalBlock(ctx, "Compile Error", err.Error())
	}

	v, err := i.Eval("main.Run()")
	if err != nil {
		return sendEvalBlock(ctx, "Runtime Error", err.Error())
	}

	return sendEvalResult(ctx, v)
}

func extractEvalCode(message string) string {
	s := strings.TrimLeft(message, " \t\n")
	for _, p := range []string{".>>", ".eval", ".ev", ">>", "eval", "ev"} {
		if strings.HasPrefix(s, p) {
			s = strings.TrimPrefix(s, p)
			return strings.TrimSpace(s)
		}
	}
	return strings.TrimSpace(s)
}

func buildEvalSetup(hasDB, hasEvent, hasRawMsg bool) string {
	var extra strings.Builder
	if hasDB {
		extra.WriteString("var db = evalctx.DB\nvar _ = db\n")
	}
	if hasEvent {
		extra.WriteString("var evt = evalctx.Event\nvar _ = evt\n")
	}
	if hasRawMsg {
		extra.WriteString("var m = evalctx.Raw\nvar _ = m\n")
	}
	return `package main

import (
	"evalctx"
)

var ctx = evalctx.Ctx
var c = evalctx.Client
var _ = ctx
var _ = c
` + extra.String()
}

func wrapEvalCode(code string) string {
	const header = `package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"strconv"
	"time"
	"regexp"
	"encoding/json"
	"io"
	"context"
	"errors"
	"sort"

	"google.golang.org/protobuf/proto"
	"github.com/jrevanaldi-ai/gowa/types"
	"github.com/jrevanaldi-ai/gowa/types/events"
	"github.com/jrevanaldi-ai/gowa/proto/waE2E"
	"github.com/jrevanaldi-ai/gowa-bot/helper"
)

var (
	_ = fmt.Sprintf
	_ = os.Getenv
	_ = exec.Command
	_ = strings.Contains
	_ = strconv.Itoa
	_ = time.Now
	_ = regexp.MustCompile
	_ = json.Marshal
	_ = io.ReadAll
	_ = context.Background
	_ = errors.New
	_ = sort.Strings
	_ = proto.String
	_ = types.ParseJID
	_ = (*events.Message)(nil)
	_ = (*waE2E.Message)(nil)
	_ = helper.CreateSimpleReply
)
`

	var body string
	if _, err := parser.ParseExpr(code); err == nil {
		body = "return " + code
	} else {
		body = code + "\n\treturn nil"
	}

	return fmt.Sprintf("%s\nfunc Run() interface{} {\n\t%s\n}\n", header, body)
}

func sendEvalResult(ctx *lib.CommandContext, v reflect.Value) error {
	var result string
	if !v.IsValid() {
		result = ""
	} else {
		raw := v.Interface()
		switch x := raw.(type) {
		case nil:
			result = ""
		case string:
			result = x
		case error:
			result = "error: " + x.Error()
		case []byte:
			result = string(x)
		default:
			result = prettyFormat(raw)
		}
	}
	if strings.TrimSpace(result) == "" {
		result = "(done, no output)"
	}
	return sendEvalBlock(ctx, "", result)
}

func prettyFormat(v interface{}) string {
	if b, err := json.MarshalIndent(v, "", "  "); err == nil {
		s := string(b)
		if s != "null" && s != "{}" && s != "" {
			return s
		}
	}
	return fmt.Sprintf("%+v", v)
}

func sendEvalBlock(ctx *lib.CommandContext, title, body string) error {
	var text string
	if title != "" {
		text = fmt.Sprintf("*%s*\n%s", title, body)
	} else {
		text = body
	}
	_, err := ctx.SendMessage(helper.CreateSimpleReply(text, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return err
}

func sendEvalUsage(ctx *lib.CommandContext) error {
	msg := "*Eval Reference*\n\n" +
		"*Mode*\n" +
		"- Expression : >> ctx.Chat.String()\n" +
		"- Statement  : pakai `return X` di akhir blok\n\n" +
		"*Variabel*\n" +
		"ctx : *lib.CommandContext\n" +
		"c   : *gowa.Client\n" +
		"evt : *events.Message  (full event WhatsApp)\n" +
		"m   : *waE2E.Message   (raw message protobuf)\n" +
		"db  : *helper.DatabaseManager\n\n" +
		"*Stdlib*\n" +
		"fmt, os, os/exec, strings, strconv,\n" +
		"time, regexp, encoding/json, io,\n" +
		"context, errors, sort\n\n" +
		"*proto*\n" +
		"String, Int32, Int64, Bool,\n" +
		"Uint32, Uint64, Float32, Float64\n\n" +
		"*types*\n" +
		"JID, ParseJID, NewJID,\n" +
		"DefaultUserServer, GroupServer,\n" +
		"NewsletterServer, HiddenUserServer,\n" +
		"BroadcastServer\n\n" +
		"*events*\n" +
		"Message, Connected, Disconnected,\n" +
		"LoggedOut, PairSuccess\n\n" +
		"*waE2E*\n" +
		"Message, ExtendedTextMessage,\n" +
		"ContextInfo, ImageMessage,\n" +
		"VideoMessage, StickerMessage,\n" +
		"ReactionMessage\n\n" +
		"*helper*\n" +
		"CreateSimpleReply, NewLogger,\n" +
		"FormatAmount, FormatFileSize\n\n" +
		"*Contoh CommandContext*\n" +
		">> ctx.Chat.String()\n" +
		">> ctx.Sender.User\n" +
		">> ctx.PushName\n" +
		">> ctx.IsGroup\n" +
		">> ctx.IsOwner\n" +
		">> ctx.Mentions\n" +
		">> ctx.ReplyMessage\n" +
		">> ctx.Args\n\n" +
		"*Contoh Event / Raw Message*\n" +
		">> evt.Info.Timestamp.String()\n" +
		">> evt.Info.PushName\n" +
		">> evt.Info.MediaType\n" +
		">> m.Conversation\n" +
		">> m.ImageMessage\n" +
		">> if m.ImageMessage != nil { return *m.ImageMessage.Caption }\n" +
		"   return \"no image\"\n\n" +
		"*Contoh Client*\n" +
		">> c.Store.ID.String()\n" +
		">> c.IsConnected()\n" +
		">> c.IsLoggedIn()\n\n" +
		"*Contoh shell exec*\n" +
		">> out, _ := exec.Command(\"grep\",\"-rn\",\"GenerateMessageID\",\"gowa-lib/send.go\").Output()\n" +
		"   return string(out)\n" +
		">> b, _ := exec.Command(\"uname\",\"-a\").Output()\n" +
		"   return string(b)\n\n" +
		"*Contoh JSON / format*\n" +
		">> j, _ := json.MarshalIndent(evt.Info, \"\", \"  \")\n" +
		"   return string(j)\n" +
		">> return fmt.Sprintf(\"%s@%s\", ctx.Sender.User, ctx.Sender.Server)\n\n" +
		"*Contoh types*\n" +
		">> jid, _ := types.ParseJID(\"62812@s.whatsapp.net\"); return jid.User\n" +
		">> return types.NewJID(\"62812\", types.DefaultUserServer).String()\n\n" +
		"*Contoh kirim pesan*\n" +
		">> ctx.SendMessage(helper.CreateSimpleReply(\"halo\", ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))\n" +
		"   return \"sent\"\n\n" +
		"*Contoh DB (kalau tersedia)*\n" +
		">> banned, _ := db.IsBanned(ctx.Sender.String(), \"user\"); return banned"
	_, err := ctx.SendMessage(helper.CreateSimpleReply(msg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	return err
}
