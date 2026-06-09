package store

import (
	"context"
	"database/sql"

	"private/ibms/internal/model"
)

// UserStore 用户数据访问接口。service 层依赖此接口而非具体实现。
type UserStore interface {
	Create(ctx context.Context, u *model.User) error
	GetByID(ctx context.Context, id int64) (*model.User, error)
	List(ctx context.Context) ([]*model.User, error)
}

type userStore struct {
	db *sql.DB
}

func (s *userStore) Create(ctx context.Context, u *model.User) error {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO users (name, email, created_at) VALUES (?, ?, ?)`,
		u.Name, u.Email, u.CreatedAt)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	u.ID = id
	return nil
}

func (s *userStore) GetByID(ctx context.Context, id int64) (*model.User, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, name, email, created_at FROM users WHERE id = ?`, id)
	var u model.User
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (s *userStore) List(ctx context.Context) ([]*model.User, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, email, created_at FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	return users, rows.Err()
}
