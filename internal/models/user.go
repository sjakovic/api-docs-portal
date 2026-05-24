package models

import (
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID                 int64     `json:"id"`
	Email              string    `json:"email"`
	PasswordHash       string    `json:"-"`
	Name               string    `json:"name"`
	Role               string    `json:"role"`
	IsActive           bool      `json:"is_active"`
	MustChangePassword bool      `json:"must_change_password"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) Create(email, password, name, role string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	mustChange := role != "admin"
	result, err := s.db.Exec(
		`INSERT INTO users (email, password_hash, name, role, must_change_password) VALUES (?, ?, ?, ?, ?)`,
		email, string(hash), name, role, mustChange,
	)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}

	id, _ := result.LastInsertId()
	return s.GetByID(id)
}

func (s *UserStore) GetByID(id int64) (*User, error) {
	u := &User{}
	err := s.db.QueryRow(
		`SELECT id, email, password_hash, name, role, is_active, must_change_password, created_at, updated_at FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.IsActive, &u.MustChangePassword, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return u, nil
}

func (s *UserStore) GetByEmail(email string) (*User, error) {
	u := &User{}
	err := s.db.QueryRow(
		`SELECT id, email, password_hash, name, role, is_active, must_change_password, created_at, updated_at FROM users WHERE email = ?`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.IsActive, &u.MustChangePassword, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return u, nil
}

func (s *UserStore) List() ([]*User, error) {
	rows, err := s.db.Query(
		`SELECT id, email, password_hash, name, role, is_active, must_change_password, created_at, updated_at FROM users ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
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

func (s *UserStore) Update(id int64, email, name, role string, isActive bool) error {
	_, err := s.db.Exec(
		`UPDATE users SET email = ?, name = ?, role = ?, is_active = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		email, name, role, isActive, id,
	)
	return err
}

func (s *UserStore) UpdatePassword(id int64, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	_, err = s.db.Exec(
		`UPDATE users SET password_hash = ?, must_change_password = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		string(hash), id,
	)
	return err
}

func (s *UserStore) Deactivate(id int64) error {
	_, err := s.db.Exec(`UPDATE users SET is_active = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, id)
	return err
}

func (s *UserStore) CheckPassword(user *User, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) == nil
}

func (s *UserStore) HasUsers() bool {
	var count int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return false
	}
	return count > 0
}

func (s *UserStore) CreateAdmin(email, password, name string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash admin password: %w", err)
	}

	result, err := s.db.Exec(
		`INSERT INTO users (email, password_hash, name, role, must_change_password) VALUES (?, ?, ?, ?, ?)`,
		email, string(hash), name, "admin", false,
	)
	if err != nil {
		return nil, fmt.Errorf("create admin: %w", err)
	}

	id, _ := result.LastInsertId()
	return s.GetByID(id)
}
