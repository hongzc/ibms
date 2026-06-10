package service

import "private/ibms/internal/store"

// Service 聚合所有业务服务，供 route 层依赖。
type Service struct {
	User      *UserService
	Dashboard *DashboardService
	BA        *BAService
	Energy    *EnergyService
	Operation *OperationService
	Workorder *WorkorderService
}

// New 基于 Store 构建 Service。
func New(s *store.Store) *Service {
	return &Service{
		User:      &UserService{users: s.User},
		Dashboard: NewDashboardService(),
		BA:        NewBAService(),
		Energy:    NewEnergyService(),
		Operation: NewOperationService(),
		Workorder: NewWorkorderService(),
	}
}
