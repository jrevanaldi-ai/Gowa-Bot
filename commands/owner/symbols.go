package owner

import (
	"reflect"

	"github.com/jrevanaldi-ai/gowa"
	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
	"github.com/jrevanaldi-ai/gowa/proto/waE2E"
	"github.com/jrevanaldi-ai/gowa/types"
	"github.com/jrevanaldi-ai/gowa/types/events"
	"google.golang.org/protobuf/proto"
)

var Symbols = map[string]map[string]reflect.Value{}

func init() {
	Symbols["google.golang.org/protobuf/proto/proto"] = map[string]reflect.Value{
		"String":  reflect.ValueOf(proto.String),
		"Int32":   reflect.ValueOf(proto.Int32),
		"Int64":   reflect.ValueOf(proto.Int64),
		"Bool":    reflect.ValueOf(proto.Bool),
		"Uint32":  reflect.ValueOf(proto.Uint32),
		"Uint64":  reflect.ValueOf(proto.Uint64),
		"Float32": reflect.ValueOf(proto.Float32),
		"Float64": reflect.ValueOf(proto.Float64),
	}

	Symbols["github.com/jrevanaldi-ai/gowa/gowa"] = map[string]reflect.Value{
		"Client": reflect.ValueOf((*gowa.Client)(nil)),
	}

	Symbols["github.com/jrevanaldi-ai/gowa/types/types"] = map[string]reflect.Value{
		"JID":               reflect.ValueOf((*types.JID)(nil)),
		"ParseJID":          reflect.ValueOf(types.ParseJID),
		"NewJID":            reflect.ValueOf(types.NewJID),
		"DefaultUserServer": reflect.ValueOf(types.DefaultUserServer),
		"GroupServer":       reflect.ValueOf(types.GroupServer),
		"NewsletterServer":  reflect.ValueOf(types.NewsletterServer),
		"HiddenUserServer":  reflect.ValueOf(types.HiddenUserServer),
		"BroadcastServer":   reflect.ValueOf(types.BroadcastServer),
	}

	Symbols["github.com/jrevanaldi-ai/gowa/proto/waE2E/waE2E"] = map[string]reflect.Value{
		"Message":             reflect.ValueOf((*waE2E.Message)(nil)),
		"ExtendedTextMessage": reflect.ValueOf((*waE2E.ExtendedTextMessage)(nil)),
		"ContextInfo":         reflect.ValueOf((*waE2E.ContextInfo)(nil)),
		"ImageMessage":        reflect.ValueOf((*waE2E.ImageMessage)(nil)),
		"VideoMessage":        reflect.ValueOf((*waE2E.VideoMessage)(nil)),
		"StickerMessage":      reflect.ValueOf((*waE2E.StickerMessage)(nil)),
		"ReactionMessage":     reflect.ValueOf((*waE2E.ReactionMessage)(nil)),
	}

	Symbols["github.com/jrevanaldi-ai/gowa/types/events/events"] = map[string]reflect.Value{
		"Message":      reflect.ValueOf((*events.Message)(nil)),
		"Connected":    reflect.ValueOf((*events.Connected)(nil)),
		"Disconnected": reflect.ValueOf((*events.Disconnected)(nil)),
		"LoggedOut":    reflect.ValueOf((*events.LoggedOut)(nil)),
		"PairSuccess":  reflect.ValueOf((*events.PairSuccess)(nil)),
	}

	Symbols["github.com/jrevanaldi-ai/gowa-bot/lib/lib"] = map[string]reflect.Value{
		"CommandContext": reflect.ValueOf((*lib.CommandContext)(nil)),
	}

	Symbols["github.com/jrevanaldi-ai/gowa-bot/helper/helper"] = map[string]reflect.Value{
		"CreateSimpleReply": reflect.ValueOf(helper.CreateSimpleReply),
		"NewLogger":         reflect.ValueOf(helper.NewLogger),
		"FormatAmount":      reflect.ValueOf(helper.FormatAmount),
		"FormatFileSize":    reflect.ValueOf(helper.FormatFileSize),
	}
}
