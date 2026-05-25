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
	GroupID     sql.NullInt64  `json:"group_id"`
	SortOrder   int            `json:"sort_order"`
	IsActive    bool           `json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	GroupName   string         `json:"-"`
}

type DocStore struct {
	db *sql.DB
}

func NewDocStore(db *sql.DB) *DocStore {
	return &DocStore{db: db}
}

const docCols = `d.id, d.name, d.slug, d.description, d.doc_type, d.content, d.external_url, d.version, d.group_id, d.sort_order, d.is_active, d.created_at, d.updated_at, COALESCE(g.name, '')`
const docJoin = `FROM docs d LEFT JOIN doc_groups g ON d.group_id = g.id`

func scanDoc(row interface{ Scan(...interface{}) error }) (*Doc, error) {
	d := &Doc{}
	err := row.Scan(&d.ID, &d.Name, &d.Slug, &d.Description, &d.DocType, &d.Content, &d.ExternalURL, &d.Version, &d.GroupID, &d.SortOrder, &d.IsActive, &d.CreatedAt, &d.UpdatedAt, &d.GroupName)
	return d, err
}

func (s *DocStore) Create(name, slug, description, docType, content, externalURL, version string, groupID int64, sortOrder int) (*Doc, error) {
	if slug == "" {
		slug = slugify(name)
	}

	var gid interface{}
	if groupID > 0 {
		gid = groupID
	}

	result, err := s.db.Exec(
		`INSERT INTO docs (name, slug, description, doc_type, content, external_url, version, group_id, sort_order) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		name, slug, nullString(description), docType, nullString(content), nullString(externalURL), nullString(version), gid, sortOrder,
	)
	if err != nil {
		return nil, fmt.Errorf("insert doc: %w", err)
	}

	id, _ := result.LastInsertId()
	return s.GetByID(id)
}

func (s *DocStore) GetByID(id int64) (*Doc, error) {
	d, err := scanDoc(s.db.QueryRow(`SELECT `+docCols+` `+docJoin+` WHERE d.id = ?`, id))
	if err != nil {
		return nil, fmt.Errorf("get doc: %w", err)
	}
	return d, nil
}

func (s *DocStore) GetBySlug(slug string) (*Doc, error) {
	d, err := scanDoc(s.db.QueryRow(`SELECT `+docCols+` `+docJoin+` WHERE d.slug = ? AND d.is_active = 1`, slug))
	if err != nil {
		return nil, fmt.Errorf("get doc by slug: %w", err)
	}
	return d, nil
}

func (s *DocStore) List() ([]*Doc, error) {
	return s.queryDocs(`SELECT ` + docCols + ` ` + docJoin + ` ORDER BY COALESCE(g.sort_order, 999999), d.sort_order, d.name`)
}

func (s *DocStore) ListForUser(userID int64) ([]*Doc, error) {
	return s.queryDocs(
		`SELECT `+docCols+` `+docJoin+`
		INNER JOIN user_doc_access uda ON d.id = uda.doc_id
		WHERE uda.user_id = ? AND d.is_active = 1
		ORDER BY COALESCE(g.sort_order, 999999), d.sort_order, d.name`, userID)
}

func (s *DocStore) ListActive() ([]*Doc, error) {
	return s.queryDocs(`SELECT ` + docCols + ` ` + docJoin + ` WHERE d.is_active = 1 ORDER BY COALESCE(g.sort_order, 999999), d.sort_order, d.name`)
}

func (s *DocStore) queryDocs(query string, args ...interface{}) ([]*Doc, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query docs: %w", err)
	}
	defer rows.Close()

	var docs []*Doc
	for rows.Next() {
		d, err := scanDoc(rows)
		if err != nil {
			return nil, fmt.Errorf("scan doc: %w", err)
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

func (s *DocStore) Update(id int64, name, slug, description, docType, content, externalURL, version string, groupID int64, sortOrder int, isActive bool) error {
	var gid interface{}
	if groupID > 0 {
		gid = groupID
	}

	_, err := s.db.Exec(
		`UPDATE docs SET name = ?, slug = ?, description = ?, doc_type = ?, content = ?, external_url = ?, version = ?, group_id = ?, sort_order = ?, is_active = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		name, slug, nullString(description), docType, nullString(content), nullString(externalURL), nullString(version), gid, sortOrder, isActive, id,
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
