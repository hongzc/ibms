package service

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Workorder struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Type      string  `json:"type"`     // 报修 | 巡检 | 投诉 | 其他
	Priority  string  `json:"priority"` // 紧急 | 重要 | 一般
	Location  string  `json:"location"`
	Creator   string  `json:"creator"`
	Assignee  string  `json:"assignee"`
	Status    string  `json:"status"` // 待派单 | 处理中 | 已完成
	CreatedAt string  `json:"createdAt"`
	Progress  float64 `json:"progress"`
	Desc      string  `json:"desc"`
}

type WorkorderStats struct {
	Total      int     `json:"total"`
	Pending    int     `json:"pending"`
	Processing int     `json:"processing"`
	Done       int     `json:"done"`
	Rate       float64 `json:"rate"`
}

type WorkorderBoard struct {
	Stats     WorkorderStats `json:"stats"`
	YearTrend []MonthCount   `json:"yearTrend"`
	ByType    []NameValue    `json:"byType"`
	List      []Workorder    `json:"list"`
}

type MonthCount struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}

// WorkorderService 工单模拟源：内存列表支持创建/派单/完成。
type WorkorderService struct {
	mu   sync.Mutex
	list []Workorder
	seq  int
}

var ErrWorkorderNotFound = fmt.Errorf("workorder not found")
var ErrWorkorderState = fmt.Errorf("invalid workorder state")

func NewWorkorderService() *WorkorderService {
	s := &WorkorderService{seq: 20}
	seeds := []struct {
		title, typ, pri, loc, creator, assignee, status string
		progress                                        float64
	}{
		{"3F 茶水间水龙头漏水", "报修", "紧急", "A栋 3F", "物业前台", "维修一组·张伟", "处理中", 60},
		{"B2 配电室季度巡检", "巡检", "一般", "B栋 B2", "运维班组", "维修二组·李强", "处理中", 35},
		{"12F 会议室空调不制冷", "报修", "重要", "A栋 12F", "星辰科技", "维修一组·王磊", "处理中", 80},
		{"大堂照明灯带闪烁", "报修", "一般", "A栋 1F", "保安队", "", "待派单", 0},
		{"电梯 2 号轿厢异响", "报修", "紧急", "B栋", "物业前台", "", "待派单", 0},
		{"P1 车库出口道闸卡顿", "报修", "重要", "地下室 P1", "停车管理", "", "待派单", 0},
		{"消防通道杂物投诉", "投诉", "重要", "C栋 5F", "租户", "综合组·赵敏", "处理中", 50},
		{"5F 烟感设备月度巡检", "巡检", "一般", "A栋 5F", "运维班组", "维修二组·李强", "已完成", 100},
		{"8F 卫生间下水道堵塞", "报修", "重要", "B栋 8F", "云启网络", "维修一组·张伟", "已完成", 100},
		{"外墙清洗作业申请", "其他", "一般", "C栋", "物业经理", "外包·洁净公司", "已完成", 100},
		{"冷却塔水质季度检测", "巡检", "一般", "屋面", "运维班组", "维修三组·陈晨", "已完成", 100},
		{"2F 商铺玻璃门破损", "报修", "紧急", "C栋 2F", "鼎盛餐饮", "维修一组·王磊", "已完成", 100},
		{"地下室积水报修", "报修", "紧急", "地下室 B2", "保安队", "维修三组·陈晨", "已完成", 100},
		{"新风滤网季度更换", "巡检", "一般", "全楼", "运维班组", "维修二组·李强", "已完成", 100},
		{"噪音扰民投诉处理", "投诉", "一般", "B栋 6F", "租户", "综合组·赵敏", "已完成", 100},
	}
	now := time.Now()
	for i, x := range seeds {
		s.list = append(s.list, Workorder{
			ID: fmt.Sprintf("WO-%s-%03d", now.AddDate(0, 0, -i/3).Format("20060102"), 15-i),
			Title: x.title, Type: x.typ, Priority: x.pri, Location: x.loc,
			Creator: x.creator, Assignee: x.assignee, Status: x.status,
			CreatedAt: now.Add(-time.Duration(i*5+2) * time.Hour).Format("2006-01-02 15:04"),
			Progress:  x.progress,
		})
	}
	return s
}

// Board 工单看板：统计 + 年度趋势 + 类型分布 + 列表。
func (s *WorkorderService) Board() WorkorderBoard {
	s.mu.Lock()
	defer s.mu.Unlock()
	var b WorkorderBoard
	byType := map[string]int{}
	for _, w := range s.list {
		b.Stats.Total++
		switch w.Status {
		case "待派单":
			b.Stats.Pending++
		case "处理中":
			b.Stats.Processing++
			// 处理中的工单进度缓慢推进，模拟现场处理
			if w.Progress < 95 {
				for i := range s.list {
					if s.list[i].ID == w.ID {
						s.list[i].Progress = jitterI(w.Progress+1, 1, w.Progress, 95)
					}
				}
			}
		case "已完成":
			b.Stats.Done++
		}
		byType[w.Type]++
	}
	// 历史基数：让统计不只有内存里这十几条
	histDone := 67
	b.Stats.Total += histDone
	b.Stats.Done += histDone
	b.Stats.Rate = round1(float64(b.Stats.Done) / float64(b.Stats.Total) * 100)

	r := rand.New(rand.NewSource(7))
	for m := 1; m <= 12; m++ {
		c := 0
		if m <= int(time.Now().Month()) {
			c = 70 + r.Intn(50)
		}
		b.YearTrend = append(b.YearTrend, MonthCount{Month: fmt.Sprintf("%d月", m), Count: c})
	}
	for _, t := range []string{"报修", "巡检", "投诉", "其他"} {
		base := map[string]int{"报修": 46, "巡检": 28, "投诉": 9, "其他": 9}[t]
		b.ByType = append(b.ByType, NameValue{Name: t, V: float64(base + byType[t])})
	}
	b.List = make([]Workorder, len(s.list))
	copy(b.List, s.list)
	return b
}

// Create 发起工单/报修，状态为待派单。
func (s *WorkorderService) Create(title, typ, location, desc string) Workorder {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	w := Workorder{
		ID: fmt.Sprintf("WO-%s-%03d", time.Now().Format("20060102"), s.seq),
		Title: title, Type: typ, Priority: "一般", Location: location,
		Creator: "管理员", Status: "待派单",
		CreatedAt: time.Now().Format("2006-01-02 15:04"), Desc: desc,
	}
	s.list = append([]Workorder{w}, s.list...)
	return w
}

// Dispatch 派单：待派单 → 处理中。
func (s *WorkorderService) Dispatch(id, assignee string) (Workorder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.list {
		if s.list[i].ID != id {
			continue
		}
		if s.list[i].Status != "待派单" {
			return Workorder{}, ErrWorkorderState
		}
		s.list[i].Status = "处理中"
		s.list[i].Assignee = assignee
		s.list[i].Progress = 10
		return s.list[i], nil
	}
	return Workorder{}, ErrWorkorderNotFound
}

// Complete 完成工单：处理中 → 已完成。
func (s *WorkorderService) Complete(id string) (Workorder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.list {
		if s.list[i].ID != id {
			continue
		}
		if s.list[i].Status != "处理中" {
			return Workorder{}, ErrWorkorderState
		}
		s.list[i].Status = "已完成"
		s.list[i].Progress = 100
		return s.list[i], nil
	}
	return Workorder{}, ErrWorkorderNotFound
}
