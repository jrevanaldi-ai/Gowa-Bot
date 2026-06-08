package webserver

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

//go:embed web
var webFS embed.FS

type Server struct {
	Addr        string
	Registry    *lib.CommandRegistry
	Client      *whatsmeow.Client
	DBManager   *helper.DatabaseManager
	JadibotMgr  *helper.JadibotSessionManager
	Activity    *helper.ActivityLog
	StartedAt   time.Time
	GetSelfMode func() bool
	GetPrefixes func() []string

	httpServer *http.Server
}

func New(addr string) *Server {
	return &Server{
		Addr:      addr,
		StartedAt: time.Now(),
	}
}

func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/info", s.handleInfo)
	mux.HandleFunc("/api/memory", s.handleMemory)
	mux.HandleFunc("/api/bot", s.handleBot)
	mux.HandleFunc("/api/commands", s.handleCommands)
	mux.HandleFunc("/api/jadibots", s.handleJadibots)
	mux.HandleFunc("/api/activity", s.handleActivity)
	mux.HandleFunc("/api/stream", s.handleStream)

	webFiles, err := fs.Sub(webFS, "web")
	if err != nil {
		return fmt.Errorf("failed to load embedded web files: %w", err)
	}
	mux.Handle("/", http.FileServer(http.FS(webFiles)))

	s.httpServer = &http.Server{
		Addr:              s.Addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.httpServer.Shutdown(shutdownCtx)
	}()

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_ = json.NewEncoder(w).Encode(data)
}

func (s *Server) buildInfo() map[string]interface{} {
	hostname, _ := os.Hostname()
	uptime := time.Since(s.StartedAt)
	return map[string]interface{}{
		"name":           "Gowa-Bot",
		"uptime_seconds": int(uptime.Seconds()),
		"started_at":     s.StartedAt.Format(time.RFC3339),
		"go_version":     runtime.Version(),
		"os":             runtime.GOOS,
		"arch":           runtime.GOARCH,
		"hostname":       hostname,
		"cpus":           runtime.NumCPU(),
		"goroutines":     runtime.NumGoroutine(),
		"pid":            os.Getpid(),
	}
}

func (s *Server) buildMemory() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	mb := func(b uint64) float64 { return float64(b) / 1024 / 1024 }
	return map[string]interface{}{
		"alloc_mb":       mb(m.Alloc),
		"total_alloc_mb": mb(m.TotalAlloc),
		"sys_mb":         mb(m.Sys),
		"heap_alloc_mb":  mb(m.HeapAlloc),
		"heap_inuse_mb":  mb(m.HeapInuse),
		"heap_idle_mb":   mb(m.HeapIdle),
		"stack_inuse_mb": mb(m.StackInuse),
		"rss_mb":         float64(getRSS()) / 1024 / 1024,
		"num_gc":         m.NumGC,
		"last_gc":        time.Unix(0, int64(m.LastGC)).Format(time.RFC3339),
	}
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.buildInfo())
}

func (s *Server) handleMemory(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.buildMemory())
}

func getRSS() uint64 {
	b, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(line, "VmRSS:") {
			var kb uint64
			_, _ = fmt.Sscanf(line, "VmRSS: %d kB", &kb)
			return kb * 1024
		}
	}
	return 0
}

func (s *Server) buildBot() map[string]interface{} {
	var phone, jid, pushName string
	connected := false
	loggedIn := false
	if s.Client != nil {
		connected = s.Client.IsConnected()
		loggedIn = s.Client.IsLoggedIn()
		if s.Client.Store != nil && s.Client.Store.ID != nil {
			phone = s.Client.Store.ID.User
			jid = s.Client.Store.ID.String()
			pushName = s.Client.Store.PushName
		}
	}

	selfMode := false
	if s.GetSelfMode != nil {
		selfMode = s.GetSelfMode()
	}

	prefixes := []string{"."}
	if s.GetPrefixes != nil {
		if p := s.GetPrefixes(); len(p) > 0 {
			prefixes = p
		}
	}

	return map[string]interface{}{
		"connected": connected,
		"logged_in": loggedIn,
		"phone":     phone,
		"jid":       jid,
		"push_name": pushName,
		"self_mode": selfMode,
		"prefixes":  prefixes,
	}
}

func (s *Server) handleBot(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.buildBot())
}

type commandDTO struct {
	Cmd       string   `json:"cmd"`
	Tag       string   `json:"tag"`
	Desc      string   `json:"desc"`
	Example   string   `json:"example"`
	OwnerOnly bool     `json:"owner_only"`
	Alias     []string `json:"alias"`
}

func (s *Server) buildCommands() map[string]interface{} {
	if s.Registry == nil {
		return map[string]interface{}{"total": 0, "commands": []interface{}{}}
	}

	all := s.Registry.GetAllCommands()
	items := make([]commandDTO, 0, len(all))
	for _, c := range all {
		alias := c.Alias
		if alias == nil {
			alias = []string{}
		}
		items = append(items, commandDTO{
			Cmd:       c.Cmd,
			Tag:       c.Tag,
			Desc:      c.Desc,
			Example:   c.Example,
			OwnerOnly: c.OwnerOnly,
			Alias:     alias,
		})
	}

	return map[string]interface{}{
		"total":    len(items),
		"commands": items,
	}
}

func (s *Server) handleCommands(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.buildCommands())
}

type jadibotDTO struct {
	ID          string `json:"id"`
	OwnerJID    string `json:"owner_jid"`
	PhoneNumber string `json:"phone_number"`
	Status      string `json:"status"`
	Running     bool   `json:"running"`
}

func (s *Server) buildJadibots() map[string]interface{} {
	empty := map[string]interface{}{
		"total":   0,
		"running": 0,
		"active":  0,
		"paused":  0,
		"stopped": 0,
		"items":   []interface{}{},
	}
	if s.DBManager == nil {
		return empty
	}

	bots, err := s.DBManager.GetAllJadibot()
	if err != nil {
		out := empty
		out["error"] = err.Error()
		return out
	}

	items := make([]jadibotDTO, 0, len(bots))
	var active, paused, stopped, running int
	for _, b := range bots {
		isRunning := false
		if s.JadibotMgr != nil {
			isRunning = s.JadibotMgr.IsRunning(b.ID)
		}
		switch b.Status {
		case "active":
			active++
		case "paused":
			paused++
		case "stopped":
			stopped++
		}
		if isRunning {
			running++
		}
		items = append(items, jadibotDTO{
			ID:          b.ID,
			OwnerJID:    b.OwnerJID,
			PhoneNumber: b.PhoneNumber,
			Status:      b.Status,
			Running:     isRunning,
		})
	}

	return map[string]interface{}{
		"total":   len(items),
		"running": running,
		"active":  active,
		"paused":  paused,
		"stopped": stopped,
		"items":   items,
	}
}

func (s *Server) handleJadibots(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.buildJadibots())
}

func (s *Server) buildActivity() map[string]interface{} {
	if s.Activity == nil {
		return map[string]interface{}{"total": 0, "items": []interface{}{}}
	}
	items := s.Activity.Recent(20)
	if items == nil {
		items = []helper.ActivityEntry{}
	}
	return map[string]interface{}{
		"total": len(items),
		"items": items,
	}
}

func (s *Server) handleActivity(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.buildActivity())
}

func (s *Server) handleStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	send := func(event string, data interface{}) bool {
		b, err := json.Marshal(data)
		if err != nil {
			return false
		}
		if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, b); err != nil {
			return false
		}
		flusher.Flush()
		return true
	}

	if !send("init", map[string]interface{}{
		"info":     s.buildInfo(),
		"memory":   s.buildMemory(),
		"bot":      s.buildBot(),
		"jadibots": s.buildJadibots(),
		"commands": s.buildCommands(),
		"activity": s.buildActivity(),
	}) {
		return
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	commandsTicker := time.NewTicker(10 * time.Second)
	defer commandsTicker.Stop()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !send("tick", map[string]interface{}{
				"info":     s.buildInfo(),
				"memory":   s.buildMemory(),
				"bot":      s.buildBot(),
				"jadibots": s.buildJadibots(),
				"activity": s.buildActivity(),
			}) {
				return
			}
		case <-commandsTicker.C:
			if !send("commands", s.buildCommands()) {
				return
			}
		}
	}
}
