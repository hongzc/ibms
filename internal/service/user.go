package service

import (
	"context"
	"errors"
	"time"

	"private/ibms/internal/model"
	"private/ibms/internal/store"
)

// ErrUserNotFound 用户不存在。
var ErrUserNotFound = errors.New("user not found")

// UserService 用户业务逻辑层，依赖 store.UserStore 接口。
type UserService struct {
	users store.UserStore
}

// CreateUserInput 创建用户入参。
type CreateUserInput struct {
	Name  string
	Email string
}

// Create 创建用户。
func (s *UserService) Create(ctx context.Context, in CreateUserInput) (*model.User, error) {
	u := &model.User{
		Name:      in.Name,
		Email:     in.Email,
		CreatedAt: time.Now(),
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// Get 按 ID 查询用户，不存在时返回 ErrUserNotFound。
func (s *UserService) Get(ctx context.Context, id int64) (*model.User, error) {
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

// List 列出所有用户。
func (s *UserService) List(ctx context.Context) ([]*model.User, error) {
	return s.users.List(ctx)
}
