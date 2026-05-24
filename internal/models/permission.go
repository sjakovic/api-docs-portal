package models

import (
	"database/sql"
	"fmt"
	"time"
)

type Permission struct {
	UserID    int64     `json:"user_id"`
	DocID     int64     `json:"doc_id"`
	GrantedAt time.Time `json:"granted_at"`
	GrantedBy int64     `json:"granted_by"`
}

type PermissionStore struct {
	db *sql.DB
}

func NewPermissionStore(db *sql.DB) *PermissionStore {
	return &PermissionStore{db: db}
}

func (s *PermissionStore) Grant(userID, docID, grantedBy int64) error {
	_, err := s.db.Exec(
		`INSERT OR IGNORE INTO user_doc_access (user_id, doc_id, granted_by) VALUES (?, ?, ?)`,
		userID, docID, grantedBy,
	)
	return err
}

func (s *PermissionStore) Revoke(userID, docID int64) error {
	_, err := s.db.Exec(`DELETE FROM user_doc_access WHERE user_id = ? AND doc_id = ?`, userID, docID)
	return err
}

func (s *PermissionStore) HasAccess(userID, docID int64) (bool, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM user_doc_access WHERE user_id = ? AND doc_id = ?`, userID, docID,
	).Scan(&count)
	return count > 0, err
}

func (s *PermissionStore) SetDocPermissions(docID int64, userIDs []int64, grantedBy int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM user_doc_access WHERE doc_id = ?`, docID); err != nil {
		return fmt.Errorf("clear permissions: %w", err)
	}

	stmt, err := tx.Prepare(`INSERT INTO user_doc_access (user_id, doc_id, granted_by) VALUES (?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	for _, uid := range userIDs {
		if _, err := stmt.Exec(uid, docID, grantedBy); err != nil {
			return fmt.Errorf("grant user %d: %w", uid, err)
		}
	}

	return tx.Commit()
}

func (s *PermissionStore) GetDocUsers(docID int64) ([]*User, error) {
	rows, err := s.db.Query(
		`SELECT u.id, u.email, u.password_hash, u.name, u.role, u.is_active, u.must_change_password, u.created_at, u.updated_at
		FROM users u
		INNER JOIN user_doc_access uda ON u.id = uda.user_id
		WHERE uda.doc_id = ?
		ORDER BY u.name`, docID,
	)
	if err != nil {
		return nil, fmt.Errorf("get doc users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u := &User{}
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.IsActive, &u.MustChangePassword, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (s *PermissionStore) GetUserDocIDs(userID int64) (map[int64]bool, error) {
	rows, err := s.db.Query(`SELECT doc_id FROM user_doc_access WHERE user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make(map[int64]bool)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids[id] = true
	}
	return ids, rows.Err()
}
