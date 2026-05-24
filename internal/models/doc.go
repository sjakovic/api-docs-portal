package models

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type Doc struct {
	ID          int64          `json:"id"`
	Name        string         `json:"name"`
	Slug        string         `json:"slug"`
	Description sql.NullString `json:"description"`
	DocType     string         `json:"doc_type"`
	Content     sql.NullString `json:"content"`
	ExternalURL sql.NullString `json:"external_url"`
	Version     sql.NullString `json:"version"`
	SortOrder   int            `json:"sort_order"`
	IsActive    bool           `json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type DocStore struct {
	db *sql.DB
}

func NewDocStore(db *sql.DB) *DocStore {
	return &DocStore{db: db}
}

func (s *DocStore) Create(name, slug, description, docType, content, externalURL, version string, sortOrder int) (*Doc, error) {
	if slug == "" {
		slug = slugify(name)
	}

	result, err := s.db.Exec(
		`INSERT INTO docs (name, slug, description, doc_type, content, external_url, version, sort_order) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		name, slug, nullString(description), docType, nullString(content), nullString(externalURL), nullString(version), sortOrder,
	)
	if err != nil {
		return nil, fmt.Errorf("insert doc: %w", err)
	}

	id, _ := result.LastInsertId()
	return s.GetByID(id)
}

func (s *DocStore) GetByID(id int64) (*Doc, error) {
	d := &Doc{}
	err := s.db.QueryRow(
		`SELECT id, name, slug, description, doc_type, content, external_url, version, sort_order, is_active, created_at, updated_at FROM docs WHERE id = ?`, id,
	).Scan(&d.ID, &d.Name, &d.Slug, &d.Description, &d.DocType, &d.Content, &d.ExternalURL, &d.Version, &d.SortOrder, &d.IsActive, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get doc: %w", err)
	}
	return d, nil
}

func (s *DocStore) GetBySlug(slug string) (*Doc, error) {
	d := &Doc{}
	err := s.db.QueryRow(
		`SELECT id, name, slug, description, doc_type, content, external_url, version, sort_order, is_active, created_at, updated_at FROM docs WHERE slug = ? AND is_active = 1`, slug,
	).Scan(&d.ID, &d.Name, &d.Slug, &d.Description, &d.DocType, &d.Content, &d.ExternalURL, &d.Version, &d.SortOrder, &d.IsActive, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get doc by slug: %w", err)
	}
	return d, nil
}

func (s *DocStore) List() ([]*Doc, error) {
	rows, err := s.db.Query(
		`SELECT id, name, slug, description, doc_type, content, external_url, version, sort_order, is_active, created_at, updated_at FROM docs ORDER BY sort_order, name`,
	)
	if err != nil {
		return nil, fmt.Errorf("list docs: %w", err)
	}
	defer rows.Close()

	var docs []*Doc
	for rows.Next() {
		d := &Doc{}
		if err := rows.Scan(&d.ID, &d.Name, &d.Slug, &d.Description, &d.DocType, &d.Content, &d.ExternalURL, &d.Version, &d.SortOrder, &d.IsActive, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan doc: %w", err)
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

func (s *DocStore) ListForUser(userID int64) ([]*Doc, error) {
	rows, err := s.db.Query(
		`SELECT d.id, d.name, d.slug, d.description, d.doc_type, d.content, d.external_url, d.version, d.sort_order, d.is_active, d.created_at, d.updated_at
		FROM docs d
		INNER JOIN user_doc_access uda ON d.id = uda.doc_id
		WHERE uda.user_id = ? AND d.is_active = 1
		ORDER BY d.sort_order, d.name`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list docs for user: %w", err)
	}
	defer rows.Close()

	var docs []*Doc
	for rows.Next() {
		d := &Doc{}
		if err := rows.Scan(&d.ID, &d.Name, &d.Slug, &d.Description, &d.DocType, &d.Content, &d.ExternalURL, &d.Version, &d.SortOrder, &d.IsActive, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan doc: %w", err)
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

func (s *DocStore) ListActive() ([]*Doc, error) {
	rows, err := s.db.Query(
		`SELECT id, name, slug, description, doc_type, content, external_url, version, sort_order, is_active, created_at, updated_at FROM docs WHERE is_active = 1 ORDER BY sort_order, name`,
	)
	if err != nil {
		return nil, fmt.Errorf("list active docs: %w", err)
	}
	defer rows.Close()

	var docs []*Doc
	for rows.Next() {
		d := &Doc{}
		if err := rows.Scan(&d.ID, &d.Name, &d.Slug, &d.Description, &d.DocType, &d.Content, &d.ExternalURL, &d.Version, &d.SortOrder, &d.IsActive, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan doc: %w", err)
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

func (s *DocStore) Update(id int64, name, slug, description, docType, content, externalURL, version string, sortOrder int, isActive bool) error {
	_, err := s.db.Exec(
		`UPDATE docs SET name = ?, slug = ?, description = ?, doc_type = ?, content = ?, external_url = ?, version = ?, sort_order = ?, is_active = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		name, slug, nullString(description), docType, nullString(content), nullString(externalURL), nullString(version), sortOrder, isActive, id,
	)
	return err
}

func (s *DocStore) Deactivate(id int64) error {
	_, err := s.db.Exec(`UPDATE docs SET is_active = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, id)
	return err
}

var nonAlphaNum = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	slug := strings.ToLower(s)
	slug = nonAlphaNum.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
