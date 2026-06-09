package store

import (
	"database/sql"

	_ "modernc.org/sqlite" // 纯 Go sqlite 驱动，注册名 "sqlite"
)

// Store 聚合所有数据访问对象，供 service 层依赖。
type Store struct {
	User UserStore
}

// New 基于已打开的数据库连接构建 Store。
func New(db *sql.DB) *Store {
	return &Store{
		User: &userStore{db: db},
	}
}

// Open 打开 sqlite 数据库连接。
func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// Migrate 建表（幂等）。
func Migrate(db *sql.DB) error {
	const schema = `
CREATE TABLE IF NOT EXISTS users (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	name       TEXT NOT NULL,
	email      TEXT NOT NULL UNIQUE,
	created_at DATETIME NOT NULL
);`
	_, err := db.Exec(schema)
	return err
}
