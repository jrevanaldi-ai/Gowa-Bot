# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
go build -o gowa-bot          # build
./gowa-bot -phone 62xxx       # run (first pairing prints an 8-char code to enter in WhatsApp → Linked Devices)
go run . -phone 62xxx         # run without building
go test ./...                 # tests (CI: .github/workflows/go.yml)
```

Flags (in `main.go`): `-phone`, `-pair`, `-db` (default `gowa-bot.db`), `-self`, `-log-level`, `-mustika-api-key`, `-ai-api-key`. Each flag has an env-var fallback prefixed `GOWA_BOT_` (see `.env.example`). `loadEnvFile()` parses `.env` at startup — no external dotenv lib.

**Go toolchain mismatch:** `go.mod` declares `go 1.26`, README says 1.21+, CI is pinned to 1.20. If `go build` fails, check the toolchain version before assuming a code issue.

**Vendored upstream:** `go.mod` has `replace github.com/jrevanaldi-ai/gowa => ./gowa-lib`. The `gowa-lib/` directory is a local fork of the Gowa WhatsApp library — edits there affect the bot, and `go mod tidy` will not pull a remote version.

## Architecture

The flow on every WhatsApp message: `gowa.Client` event → `BotClient.EventHandler` (`client/bot_client.go`) → `processMessage` → registry lookup → handler in `commands/<category>/`.

### Core types (`lib/`)
- `CommandRegistry` — map of `cmd name → (metadata, handler)`. Aliases register the same handler under each alias key.
- `CommandMetadata` — `Cmd`, `Tag` (category, used by `.menu` grouping), `Desc`, `Example`, `Hidden`, `OwnerOnly`, `Alias`.
- `CommandContext` — what every handler receives: `Client`, `BotClient`, `JadibotSessionManager`, `Sender`/`Chat` JIDs, `IsOwner`, `IsGroup`, `Args`, `MessageID`, `ReplyMessage`, `Mentions`, plus `EphemeralWrapper` and a `SendMessage` helper that automatically applies ephemeral wrapping in groups.
- `BotClientInterface` / `JadibotSessionManagerInterface` — interfaces so command code in `commands/owner/` etc. doesn't import `client/` (which would be a cycle since `client/` imports `commands/owner` for `ParseExecCommand`).

### Message handling rules (`client/bot_client.go`)
- Prefix list is mutable per-bot (`SetPrefixes`); default `["."]`. Owners can call commands **without** prefix (unless message `IsFromMe` — anti-loop), and `$<cmd>` is reserved for `exec`.
- `processMessage` recursively unwraps `EditedMessage`, `EphemeralMessage`, `ViewOnceMessage[V2]`, and `DocumentWithCaptionMessage` so handlers see the inner payload.
- Self-mode + IsFromMe gates: non-owner self-messages are always dropped; in public mode (default) IsFromMe is dropped unless owner; toggle via `.setmode`.
- Ban checks (`DBManager.IsBanned`) run before dispatch for non-owners — both group JID and user JID.
- Special case: any message containing the substring `"lune"` (case-insensitive) routes to the `lune` AI handler even without a command prefix.

### Adding a command
1. Create `commands/<category>/<name>.go` defining `<Name>Metadata *lib.CommandMetadata` and `<Name>Handler(ctx *lib.CommandContext) error`.
2. Register in `main.go::registerCommands` — **must be added there**, no autoload.
3. Send replies via `ctx.SendMessage(helper.CreateSimpleReply(text, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))`. `CreateSimpleReply` builds the quoted-reply `ExtendedTextMessage` correctly.

### Jadibot (multi-tenant sub-bots)
`JadibotSessionManager` (`helper/session_manager.go`) lets users pair their own WhatsApp number as a child bot under the main bot. Each child:
- Gets its own SQLite store at `sessions/jadibot_<uuid>/jadibot.db` (separate from `gowa-bot.db`).
- Shares the same `CommandRegistry` as the main bot.
- Is built through a `ClientFactory` closure injected from `main.go` (this avoids a `helper` ↔ `client` import cycle).
- States `active` / `paused` / `stopped` are persisted in the `jadibots` table; active ones auto-resume at startup.
- Auto-reconnects up to 5× on disconnect, stops on logout.

### Database (`helper/database.go`)
Single SQLite file (`gowa-bot.db` by default) holds the Gowa WhatsApp session tables **and** app tables: `jadibots`, `banned`, `donations`. Schema is created idempotently in `createTables()`. Connection uses `?_foreign_keys=on`.

### Optional integrations (gated by env/flag)
- `GOWA_BOT_MUSTIKA_API_KEY` → enables `general.SetMustikaPayAPIKey` (QRIS donations via MustikaPay).
- `GOWA_BOT_AI_API_KEY` → enables `helper.NewAIService` (Claude AI for `.lune`).
Both are opt-in; main bootstraps cleanly when unset.

### Eval (`commands/owner/eval.go`)
Owner-only `>>` runs arbitrary Go through **yaegi** interpreter. `commands/owner/symbols.go` registers types/functions exposed to evaluated code; `ctx`, `c` (client), and `db` are pre-bound. Treat this as a privileged debug surface — never relax the `OwnerOnly` gate.

## Gotchas

- `gowa-bot.db` in the repo root is a real session DB — do not commit, do not delete while bot is running. `.gitignore` excludes it.
- The `lib/dispatcher.go` `Dispatcher` type is *not* used by the current message flow — dispatch happens inline in `BotClient.processMessage`. Don't wire new logic through `Dispatcher` expecting it to run.
- `client/` imports `commands/owner` (for `ParseExecCommand`). Don't make `commands/owner` import `client/` — use `lib.BotClientInterface` instead.
- Commands sent from main bot are evaluated against the main bot's owner list; jadibot child clients are created with empty owners and rely on `isOwner` matching the child's own `Store.ID.User`.
