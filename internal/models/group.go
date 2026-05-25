package models

import (
	"database/sql"
	"fmt"
	"time"
)

type DocGroup struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

type DocGroupStore struct {
	db *sql.DB
}

func NewDocGroupStore(db *sql.DB) *DocGroupStore {
	return &DocGroupStore{db: db}
}

func (s *DocGroupStore) Create(name string, sortOrder int) (*DocGroup, error) {
	result, err := s.db.Exec(
		`INSERT INTO doc_groups (name, sort_order) VALUES (?, ?)`,
		name, sortOrder,
	)
	if err != nil {
		return nil, fmt.Errorf("insert group: %w", err)
	}
	id, _ := result.LastInsertId()
	return s.GetByID(id)
}

func (s *DocGroupStore) GetByID(id int64) (*DocGroup, error) {
	g := &DocGroup{}
	err := s.db.QueryRow(
		`SELECT id, name, sort_order, created_at FROM doc_groups WHERE id = ?`, id,
	).Scan(&g.ID, &g.Name, &g.SortOrder, &g.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get group: %w", err)
	}
	return g, nil
}

func (s *DocGroupStore) List() ([]*DocGroup, error) {
	rows, err := s.db.Query(
		`SELECT id, name, sort_order, created_at FROM doc_groups ORDER BY sort_order, name`,
	)
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	defer rows.Close()

	var groups []*DocGroup
	for rows.Next() {
		g := &DocGroup{}
		if err := rows.Scan(&g.ID, &g.Name, &g.SortOrder, &g.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan group: %w", err)
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

func (s *DocGroupStore) Update(id int64, name string, sortOrder int) error {
	_, err := s.db.Exec(
		`UPDATE doc_groups SET name = ?, sort_order = ? WHERE id = ?`,
		name, sortOrder, id,
	)
	return err
}

func (s *DocGroupStore) Delete(id int64) error {
	_, err := s.db.Exec(`UPDATE docs SET group_id = NULL WHERE group_id = ?`, id)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`DELETE FROM doc_groups WHERE id = ?`, id)
	return err
}
