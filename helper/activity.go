package helper

import (
	"sync"
	"time"
)

type ActivityEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Command   string    `json:"command"`
	Sender    string    `json:"sender"`
	PushName  string    `json:"push_name"`
	ChatType  string    `json:"chat_type"`
	IsOwner   bool      `json:"is_owner"`
}

type ActivityLog struct {
	mu      sync.RWMutex
	entries []ActivityEntry
	maxSize int
}

func NewActivityLog(maxSize int) *ActivityLog {
	if maxSize <= 0 {
		maxSize = 50
	}
	return &ActivityLog{
		entries: make([]ActivityEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

func (a *ActivityLog) Push(entry ActivityEntry) {
	if a == nil {
		return
	}
	entry.Timestamp = time.Now()
	a.mu.Lock()
	defer a.mu.Unlock()
	a.entries = append(a.entries, entry)
	if len(a.entries) > a.maxSize {
		a.entries = a.entries[len(a.entries)-a.maxSize:]
	}
}

func (a *ActivityLog) Recent(limit int) []ActivityEntry {
	if a == nil {
		return nil
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	n := len(a.entries)
	if limit <= 0 || limit > n {
		limit = n
	}
	out := make([]ActivityEntry, limit)
	for i := 0; i < limit; i++ {
		out[i] = a.entries[n-1-i]
	}
	return out
}

func (a *ActivityLog) Len() int {
	if a == nil {
		return 0
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.entries)
}
