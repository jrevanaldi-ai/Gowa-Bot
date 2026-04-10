package helper

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)


type JadibotStatus string

const (
	StatusActive   JadibotStatus = "active"
	StatusPaused   JadibotStatus = "paused"
	StatusStopped  JadibotStatus = "stopped"
)


type JadibotInfo = lib.JadibotInfo


type DatabaseManager struct {
	DB *sql.DB
}


func NewDatabaseManager(dbPath string) (*DatabaseManager, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}


	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	manager := &DatabaseManager{DB: db}


	if err := manager.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return manager, nil
}


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

	CREATE TABLE IF NOT EXISTS banned (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL CHECK(type IN ('group', 'user')),
		jid TEXT NOT NULL UNIQUE,
		reason TEXT,
		banned_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		banned_by TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_jadibots_owner ON jadibots(owner_jid);
	CREATE INDEX IF NOT EXISTS idx_jadibots_status ON jadibots(status);
	CREATE INDEX IF NOT EXISTS idx_banned_type ON banned(type);
	CREATE INDEX IF NOT EXISTS idx_banned_jid ON banned(jid);
	`

	_, err := m.DB.Exec(query)
	return err
}


func (m *DatabaseManager) CreateJadibot(info JadibotInfo) error {
	query := `
	INSERT INTO jadibots (id, owner_jid, phone_number, session_path, status, created_at)
	VALUES (?, ?, ?, ?, 'stopped', ?)
	`

	_, err := m.DB.Exec(query, info.ID, info.OwnerJID, info.PhoneNumber, info.SessionPath, time.Now())
	return err
}


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


func (m *DatabaseManager) DeleteJadibot(id string) error {
	query := `DELETE FROM jadibots WHERE id = ?`
	_, err := m.DB.Exec(query, id)
	return err
}


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


func (m *DatabaseManager) Close() error {
	return m.DB.Close()
}


// BanJID menambahkan grup atau user ke daftar banned
func (m *DatabaseManager) BanJID(jid string, banType string, reason string, bannedBy string) error {
	query := `
	INSERT OR REPLACE INTO banned (id, type, jid, reason, banned_at, banned_by)
	VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, ?)
	`

	id := fmt.Sprintf("%s_%s", banType, jid)
	_, err := m.DB.Exec(query, id, banType, jid, reason, bannedBy)
	return err
}


// UnbanJID menghapus grup atau user dari daftar banned
func (m *DatabaseManager) UnbanJID(jid string, banType string) error {
	query := `DELETE FROM banned WHERE type = ? AND jid = ?`
	_, err := m.DB.Exec(query, banType, jid)
	return err
}


// IsBanned mengecek apakah grup atau user sedang di-banned
func (m *DatabaseManager) IsBanned(jid string, banType string) (bool, error) {
	query := `SELECT COUNT(*) FROM banned WHERE type = ? AND jid = ?`
	var count int
	err := m.DB.QueryRow(query, banType, jid).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}


// GetBannedList mengambil daftar semua yang di-banned
func (m *DatabaseManager) GetBannedList(banType string) ([]map[string]string, error) {
	var query string
	var rows *sql.Rows
	var err error

	if banType != "" {
		query = `SELECT id, type, jid, reason, banned_at, banned_by FROM banned WHERE type = ? ORDER BY banned_at DESC`
		rows, err = m.DB.Query(query, banType)
	} else {
		query = `SELECT id, type, jid, reason, banned_at, banned_by FROM banned ORDER BY banned_at DESC`
		rows, err = m.DB.Query(query)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]string
	for rows.Next() {
		var id, bType, jid, bannedBy string
		var reason sql.NullString
		var bannedAt string

		err := rows.Scan(&id, &bType, &jid, &reason, &bannedAt, &bannedBy)
		if err != nil {
			return nil, err
		}

		reasonText := ""
		if reason.Valid {
			reasonText = reason.String
		}

		result = append(result, map[string]string{
			"type":     bType,
			"jid":      jid,
			"reason":   reasonText,
			"bannedAt": bannedAt,
			"bannedBy": bannedBy,
		})
	}

	return result, nil
}


// GetBannedCount menghitung jumlah yang di-banned berdasarkan tipe
func (m *DatabaseManager) GetBannedCount(banType string) (int, error) {
	var query string
	var count int

	if banType != "" {
		query = `SELECT COUNT(*) FROM banned WHERE type = ?`
		err := m.DB.QueryRow(query, banType).Scan(&count)
		if err != nil {
			return 0, err
		}
	} else {
		query = `SELECT COUNT(*) FROM banned`
		err := m.DB.QueryRow(query).Scan(&count)
		if err != nil {
			return 0, err
		}
	}

	return count, nil
}
