package helper

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)

// JadibotStatus representasi status jadibot
type JadibotStatus string

const (
	StatusActive   JadibotStatus = "active"
	StatusPaused   JadibotStatus = "paused"
	StatusStopped  JadibotStatus = "stopped"
)

// JadibotInfo menyimpan informasi jadibot - menggunakan lib.JadibotInfo
type JadibotInfo = lib.JadibotInfo

// DatabaseManager mengelola database SQLite untuk jadibot
type DatabaseManager struct {
	DB *sql.DB
}

// NewDatabaseManager membuat instance database manager baru
func NewDatabaseManager(dbPath string) (*DatabaseManager, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	manager := &DatabaseManager{DB: db}

	// Create tables
	if err := manager.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return manager, nil
}

// createTables membuat tabel yang diperlukan
func (m *DatabaseManager) createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS jadibots (
		id TEXT PRIMARY KEY,
		owner_jid TEXT NOT NULL,
		phone_number TEXT NOT NULL,
		session_path TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'stopped',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		started_at DATETIME,
		last_active_at DATETIME
	);
	
	CREATE INDEX IF NOT EXISTS idx_jadibots_owner ON jadibots(owner_jid);
	CREATE INDEX IF NOT EXISTS idx_jadibots_status ON jadibots(status);
	`

	_, err := m.DB.Exec(query)
	return err
}

// CreateJadibot membuat jadibot baru
func (m *DatabaseManager) CreateJadibot(info JadibotInfo) error {
	query := `
	INSERT INTO jadibots (id, owner_jid, phone_number, session_path, status, created_at)
	VALUES (?, ?, ?, ?, 'stopped', ?)
	`

	_, err := m.DB.Exec(query, info.ID, info.OwnerJID, info.PhoneNumber, info.SessionPath, time.Now())
	return err
}

// GetJadibot mendapatkan jadibot berdasarkan ID
func (m *DatabaseManager) GetJadibot(id string) (*JadibotInfo, error) {
	query := `
	SELECT id, owner_jid, phone_number, session_path, status, created_at, started_at, last_active_at
	FROM jadibots WHERE id = ?
	`

	var info JadibotInfo
	var createdAt, startedAt, lastActiveAt sql.NullTime
	err := m.DB.QueryRow(query, id).Scan(
		&info.ID, &info.OwnerJID, &info.PhoneNumber, &info.SessionPath,
		&info.Status, &createdAt, &startedAt, &lastActiveAt,
	)

	if err != nil {
		return nil, err
	}

	// Convert sql.NullTime to interface{}
	if createdAt.Valid {
		info.CreatedAt = createdAt.Time
	}
	if startedAt.Valid {
		info.StartedAt = startedAt.Time
	}
	if lastActiveAt.Valid {
		info.LastActiveAt = lastActiveAt.Time
	}

	return &info, nil
}

// GetJadibotByOwner mendapatkan jadibot berdasarkan owner JID
func (m *DatabaseManager) GetJadibotByOwner(ownerJID string) ([]JadibotInfo, error) {
	query := `
	SELECT id, owner_jid, phone_number, session_path, status, created_at, started_at, last_active_at
	FROM jadibots WHERE owner_jid = ? ORDER BY created_at DESC
	`

	rows, err := m.DB.Query(query, ownerJID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jadibots []JadibotInfo
	for rows.Next() {
		var info JadibotInfo
		var createdAt, startedAt, lastActiveAt sql.NullTime
		err := rows.Scan(
			&info.ID, &info.OwnerJID, &info.PhoneNumber, &info.SessionPath,
			&info.Status, &createdAt, &startedAt, &lastActiveAt,
		)
		if err != nil {
			return nil, err
		}
		
		// Convert sql.NullTime to interface{}
		if createdAt.Valid {
			info.CreatedAt = createdAt.Time
		}
		if startedAt.Valid {
			info.StartedAt = startedAt.Time
		}
		if lastActiveAt.Valid {
			info.LastActiveAt = lastActiveAt.Time
		}
		
		jadibots = append(jadibots, info)
	}

	return jadibots, nil
}

// GetAllJadibot mendapatkan semua jadibot
func (m *DatabaseManager) GetAllJadibot() ([]JadibotInfo, error) {
	query := `
	SELECT id, owner_jid, phone_number, session_path, status, created_at, started_at, last_active_at
	FROM jadibots ORDER BY created_at DESC
	`

	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jadibots []JadibotInfo
	for rows.Next() {
		var info JadibotInfo
		var createdAt, startedAt, lastActiveAt sql.NullTime
		err := rows.Scan(
			&info.ID, &info.OwnerJID, &info.PhoneNumber, &info.SessionPath,
			&info.Status, &createdAt, &startedAt, &lastActiveAt,
		)
		if err != nil {
			return nil, err
		}
		
		// Convert sql.NullTime to interface{}
		if createdAt.Valid {
			info.CreatedAt = createdAt.Time
		}
		if startedAt.Valid {
			info.StartedAt = startedAt.Time
		}
		if lastActiveAt.Valid {
			info.LastActiveAt = lastActiveAt.Time
		}
		
		jadibots = append(jadibots, info)
	}

	return jadibots, nil
}

// UpdateJadibotStatus update status jadibot
func (m *DatabaseManager) UpdateJadibotStatus(id string, status JadibotStatus) error {
	query := `
	UPDATE jadibots 
	SET status = ?, 
	    started_at = CASE WHEN ? = 'active' THEN CURRENT_TIMESTAMP ELSE started_at END,
	    last_active_at = CURRENT_TIMESTAMP
	WHERE id = ?
	`

	_, err := m.DB.Exec(query, status, status, id)
	return err
}

// DeleteJadibot menghapus jadibot
func (m *DatabaseManager) DeleteJadibot(id string) error {
	query := `DELETE FROM jadibots WHERE id = ?`
	_, err := m.DB.Exec(query, id)
	return err
}

// GetActiveJadibot mendapatkan semua jadibot yang active
func (m *DatabaseManager) GetActiveJadibot() ([]JadibotInfo, error) {
	query := `
	SELECT id, owner_jid, phone_number, session_path, status, created_at, started_at, last_active_at
	FROM jadibots WHERE status = 'active' ORDER BY created_at DESC
	`

	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jadibots []JadibotInfo
	for rows.Next() {
		var info JadibotInfo
		var createdAt, startedAt, lastActiveAt sql.NullTime
		err := rows.Scan(
			&info.ID, &info.OwnerJID, &info.PhoneNumber, &info.SessionPath,
			&info.Status, &createdAt, &startedAt, &lastActiveAt,
		)
		if err != nil {
			return nil, err
		}
		
		// Convert sql.NullTime to interface{}
		if createdAt.Valid {
			info.CreatedAt = createdAt.Time
		}
		if startedAt.Valid {
			info.StartedAt = startedAt.Time
		}
		if lastActiveAt.Valid {
			info.LastActiveAt = lastActiveAt.Time
		}
		
		jadibots = append(jadibots, info)
	}

	return jadibots, nil
}

// Close menutup koneksi database
func (m *DatabaseManager) Close() error {
	return m.DB.Close()
}
