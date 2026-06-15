package store

import (
	"context"
	"database/sql"
	"time"

	"private/ibms/internal/model"
)

// ScadaStore 组态工程数据访问接口。
type ScadaStore interface {
	Create(ctx context.Context, name, graph string) (int64, error)
	Save(ctx context.Context, id int64, name, graph string) error
	Get(ctx context.Context, id int64) (*model.ScadaProject, error)
	List(ctx context.Context) ([]*model.ScadaProject, error) // 不含 graph，仅列表元信息
	Delete(ctx context.Context, id int64) error
	SetPublished(ctx context.Context, id int64) error // 中台「下发到大屏」
	GetPublishedID(ctx context.Context) (int64, error) // 0 表示未下发
}

type scadaStore struct {
	db *sql.DB
}

func (s *scadaStore) Create(ctx context.Context, name, graph string) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO scada_projects (name, graph, updated_at) VALUES (?, ?, ?)`,
		name, graph, time.Now())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *scadaStore) Save(ctx context.Context, id int64, name, graph string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE scada_projects SET name = ?, graph = ?, updated_at = ? WHERE id = ?`,
		name, graph, time.Now(), id)
	return err
}

func (s *scadaStore) Get(ctx context.Context, id int64) (*model.ScadaProject, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, name, graph, updated_at FROM scada_projects WHERE id = ?`, id)
	var p model.ScadaProject
	var graph string
	if err := row.Scan(&p.ID, &p.Name, &graph, &p.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if graph != "" {
		p.Graph = []byte(graph)
	}
	return &p, nil
}

func (s *scadaStore) List(ctx context.Context) ([]*model.ScadaProject, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, updated_at FROM scada_projects ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*model.ScadaProject
	for rows.Next() {
		var p model.ScadaProject
		if err := rows.Scan(&p.ID, &p.Name, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, &p)
	}
	return out, rows.Err()
}

func (s *scadaStore) Delete(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM scada_projects WHERE id = ?`, id)
	return err
}

// SetPublished 标记某工程为「已下发到大屏」（单例：k 恒为 1）。
func (s *scadaStore) SetPublished(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO scada_publish (k, project_id) VALUES (1, ?)
		 ON CONFLICT(k) DO UPDATE SET project_id = excluded.project_id`, id)
	return err
}

// GetPublishedID 返回已下发的工程 id，未下发返回 0。
func (s *scadaStore) GetPublishedID(ctx context.Context) (int64, error) {
	row := s.db.QueryRowContext(ctx, `SELECT project_id FROM scada_publish WHERE k = 1`)
	var id int64
	if err := row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return id, nil
}
