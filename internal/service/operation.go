package service

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type OperationOverview struct {
	TotalArea     float64             `json:"totalArea"`
	RentRate      float64             `json:"rentRate"`
	AvgPrice      float64             `json:"avgPrice"`
	LeasableArea  float64             `json:"leasableArea"`
	BillingRate   float64             `json:"billingRate"`
	Buildings     []BuildingStat      `json:"buildings"`
}

type BuildingStat struct {
	Name          string  `json:"name"`
	Area          float64 `json:"area"`
	RentRate      float64 `json:"rentRate"`
	AvgPrice      float64 `json:"avgPrice"`
	LeasableArea  float64 `json:"leasableArea"`
	Rooms         int     `json:"rooms"`
	LeasableRooms int     `json:"leasableRooms"`
}

type Room struct {
	ID          string  `json:"id"`
	Building    string  `json:"building"`
	Floor       int     `json:"floor"`
	Name        string  `json:"name"`
	Area        float64 `json:"area"`
	Status      string  `json:"status"` // 已租 | 空置 | 欠费 | 过期
	Type        string  `json:"type"`
	Tenant      string  `json:"tenant"`
	Price       float64 `json:"price"`
	ContractEnd string  `json:"contractEnd"`
	Debt        float64 `json:"debt"`
}

type Lease struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Tenant       string  `json:"tenant"`
	Contact      string  `json:"contact"`
	Room         string  `json:"room"`
	Start        string  `json:"start"`
	End          string  `json:"end"`
	MonthlyRent  float64 `json:"monthlyRent"`
	PaidProgress float64 `json:"paidProgress"`
	TermProgress float64 `json:"termProgress"`
	Status       string  `json:"status"` // 履约中 | 即将到期 | 已逾期
	ExpireDays   int     `json:"expireDays"`
}

// OperationService 运营数据模拟源：房态与租约在内存中固定生成。
type OperationService struct {
	mu     sync.Mutex
	rooms  []Room
	leases []Lease
}

var opTenants = []string{
	"星辰科技有限公司", "云启网络科技", "恒达贸易公司", "悦容美容会所", "鼎盛餐饮集团",
	"华仁律师事务所", "蓝海广告传媒", "天驰物流股份", "明朗教育咨询", "卓远建筑设计院",
	"嘉禾人力资源", "睿智金融服务",
}

func NewOperationService() *OperationService {
	s := &OperationService{}
	types := []string{"办公", "办公", "办公", "商业", "餐饮"}
	r := rand.New(rand.NewSource(42)) // 固定种子：每次启动房态一致
	now := time.Now()
	for bi, b := range []struct {
		name   string
		floors int
		per    int
	}{{"A 栋", 6, 6}, {"B 栋", 5, 6}, {"C 栋", 4, 5}} {
		for f := 1; f <= b.floors; f++ {
			floor := f * 2 // 取偶数层模拟高层
			for k := 1; k <= b.per; k++ {
				status := "已租"
				switch v := r.Float64(); {
				case v < 0.18:
					status = "空置"
				case v < 0.24:
					status = "欠费"
				case v < 0.28:
					status = "过期"
				}
				room := Room{
					ID: fmt.Sprintf("%c-%02d-%02d", 'A'+bi, floor, k), Building: b.name, Floor: floor,
					Name: fmt.Sprintf("%d%02d", floor, k), Area: float64(120 + r.Intn(40)*10),
					Status: status, Type: types[r.Intn(len(types))],
					Price: float64(35+r.Intn(20)) / 10,
				}
				if status != "空置" {
					room.Tenant = opTenants[r.Intn(len(opTenants))]
					end := now.AddDate(0, r.Intn(30)-3, r.Intn(28))
					room.ContractEnd = end.Format("2006-01-02")
					if status == "欠费" {
						room.Debt = float64(8000 + r.Intn(40000))
					}
				}
				s.rooms = append(s.rooms, room)
			}
		}
	}
	for i, t := range opTenants {
		start := now.AddDate(0, -6-i, 0)
		end := start.AddDate(2, 0, 0)
		expire := int(time.Until(end).Hours() / 24)
		status := "履约中"
		if expire < 0 {
			status = "已逾期"
		} else if expire <= 90 {
			status = "即将到期"
		}
		term := float64(now.Sub(start)) / float64(end.Sub(start)) * 100
		paid := 100.0
		if status != "履约中" {
			paid = float64(70 + i*2%30)
		}
		s.leases = append(s.leases, Lease{
			ID: fmt.Sprintf("HT-2025-%03d", i+1), Name: t + "租赁合同", Tenant: t,
			Contact: fmt.Sprintf("139%08d", 1000000+i*137913), Room: s.rooms[i*5%len(s.rooms)].Building + " " + s.rooms[i*5%len(s.rooms)].Name,
			Start: start.Format("2006-01-02"), End: end.Format("2006-01-02"),
			MonthlyRent: float64(20000 + i*4500), PaidProgress: paid,
			TermProgress: float64(int(term)), Status: status, ExpireDays: expire,
		})
	}
	return s
}

// Overview 楼宇租赁统计（从房态实时汇总）。
func (s *OperationService) Overview() OperationOverview {
	s.mu.Lock()
	defer s.mu.Unlock()
	type acc struct {
		area, rented, leasable, priceSum float64
		rooms, leasableRooms, rentedN    int
	}
	byB := map[string]*acc{}
	var order []string
	total := acc{}
	for _, r := range s.rooms {
		a, ok := byB[r.Building]
		if !ok {
			a = &acc{}
			byB[r.Building] = a
			order = append(order, r.Building)
		}
		a.area += r.Area
		a.rooms++
		total.area += r.Area
		total.rooms++
		if r.Status == "空置" {
			a.leasable += r.Area
			a.leasableRooms++
			total.leasable += r.Area
		} else {
			a.rented += r.Area
			a.priceSum += r.Price
			a.rentedN++
			total.rented += r.Area
			total.priceSum += r.Price
			total.rentedN++
		}
	}
	out := OperationOverview{
		TotalArea:    total.area,
		RentRate:     round1(total.rented / total.area * 100),
		AvgPrice:     round1(total.priceSum / float64(max(total.rentedN, 1))),
		LeasableArea: total.leasable,
		BillingRate:  92.3,
	}
	for _, name := range order {
		a := byB[name]
		out.Buildings = append(out.Buildings, BuildingStat{
			Name: name, Area: a.area,
			RentRate: round1(a.rented / a.area * 100), AvgPrice: round1(a.priceSum / float64(max(a.rentedN, 1))),
			LeasableArea: a.leasable, Rooms: a.rooms, LeasableRooms: a.leasableRooms,
		})
	}
	return out
}

// Rooms 房态列表，可按楼栋筛选（空串返回全部）。
func (s *OperationService) Rooms(building string) []Room {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []Room
	for _, r := range s.rooms {
		if building == "" || r.Building == building {
			out = append(out, r)
		}
	}
	return out
}

// Leases 租约列表。
func (s *OperationService) Leases() []Lease {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Lease, len(s.leases))
	copy(out, s.leases)
	return out
}

func round1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}
