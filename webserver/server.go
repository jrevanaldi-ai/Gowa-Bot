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

	"github.com/jrevanaldi-ai/gowa"
	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

//go:embed web
var webFS embed.FS

type Server struct {
	Addr       string
	Registry   *lib.CommandRegistry
	Client     *gowa.Client
	DBManager  *helper.DatabaseManager
	JadibotMgr *helper.JadibotSessionManager
	StartedAt  time.Time
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

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()
	uptime := time.Since(s.StartedAt)

	writeJSON(w, map[string]interface{}{
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
	})
}

func (s *Server) handleMemory(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	mb := func(b uint64) float64 { return float64(b) / 1024 / 1024 }

	writeJSON(w, map[string]interface{}{
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
	})
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

func (s *Server) handleBot(w http.ResponseWriter, r *http.Request) {
	var phone, jid string
	var pushName string
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

	writeJSON(w, map[string]interface{}{
		"connected": connected,
		"logged_in": loggedIn,
		"phone":     phone,
		"jid":       jid,
		"push_name": pushName,
		"self_mode": selfMode,
		"prefixes":  prefixes,
	})
}

type commandDTO struct {
	Cmd       string   `json:"cmd"`
	Tag       string   `json:"tag"`
	Desc      string   `json:"desc"`
	Example   string   `json:"example"`
	OwnerOnly bool     `json:"owner_only"`
	Alias     []string `json:"alias"`
}

func (s *Server) handleCommands(w http.ResponseWriter, r *http.Request) {
	if s.Registry == nil {
		writeJSON(w, map[string]interface{}{"total": 0, "commands": []interface{}{}})
		return
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

	writeJSON(w, map[string]interface{}{
		"total":    len(items),
		"commands": items,
	})
}

type jadibotDTO struct {
	ID          string `json:"id"`
	OwnerJID    string `json:"owner_jid"`
	PhoneNumber string `json:"phone_number"`
	Status      string `json:"status"`
	Running     bool   `json:"running"`
}

func (s *Server) handleJadibots(w http.ResponseWriter, r *http.Request) {
	if s.DBManager == nil {
		writeJSON(w, map[string]interface{}{"total": 0, "items": []interface{}{}})
		return
	}

	bots, err := s.DBManager.GetActiveJadibot()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]interface{}{"error": err.Error()})
		return
	}

	items := make([]jadibotDTO, 0, len(bots))
	for _, b := range bots {
		running := false
		if s.JadibotMgr != nil {
			running = s.JadibotMgr.IsRunning(b.ID)
		}
		items = append(items, jadibotDTO{
			ID:          b.ID,
			OwnerJID:    b.OwnerJID,
			PhoneNumber: b.PhoneNumber,
			Status:      b.Status,
			Running:     running,
		})
	}

	writeJSON(w, map[string]interface{}{
		"total": len(items),
		"items": items,
	})
}
