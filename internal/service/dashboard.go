package service

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
)

// Dashboard 与前端 initialData() 同构的大屏聚合数据。
type Dashboard struct {
	Weather     Weather       `json:"weather"`
	Env         Env           `json:"env"`
	Devices     Devices       `json:"devices"`
	Energy      []EnergyItem  `json:"energy"`
	Parking     Parking       `json:"parking"`
	Workorder   WorkorderSummary `json:"workorder"`
	Room        MachineRoom      `json:"room"`
	Broadcast   Broadcast     `json:"broadcast"`
	AlarmTrend  []TrendPoint  `json:"alarmTrend"`
	SecurityCat []SecurityCat `json:"securityCat"`
}

type Weather struct {
	City string  `json:"city"`
	Temp float64 `json:"temp"`
	Desc string  `json:"desc"`
	Icon string  `json:"icon"`
}

type Env struct {
	Temp     float64 `json:"temp"`
	Humidity float64 `json:"humidity"`
	PM25     float64 `json:"pm25"`
	CO2      float64 `json:"co2"`
}

type Devices struct {
	Online  int `json:"online"`
	Offline int `json:"offline"`
	Fault   int `json:"fault"`
}

type EnergyItem struct {
	Name  string  `json:"name"`
	V     float64 `json:"v"`
	Color string  `json:"color"`
}

type Parking struct {
	Total int `json:"total"`
	Used  int `json:"used"`
	Free  int `json:"free"`
}

type WorkorderSummary struct {
	Repair  int `json:"repair"`
	Inspect int `json:"inspect"`
	Done    int `json:"done"`
	Undone  int `json:"undone"`
}

type MachineRoom struct {
	Temp     float64 `json:"temp"`
	Humidity float64 `json:"humidity"`
	Power    string  `json:"power"`
	Status   string  `json:"status"`
}

type Broadcast struct {
	Zones   int `json:"zones"`
	Playing int `json:"playing"`
	Volume  int `json:"volume"`
}

type TrendPoint struct {
	T         string `json:"t"`
	Intrusion int    `json:"入侵"`
	Fire      int    `json:"消防"`
	Device    int    `json:"设备"`
}

type SecurityCat struct {
	Name  string `json:"name"`
	V     int    `json:"v"`
	Color string `json:"color"`
}

type Alarm struct {
	ID    string `json:"id"`
	Time  string `json:"time"`
	Dev   string `json:"dev"`
	Type  string `json:"type"`
	Level string `json:"level"`
	Desc  string `json:"desc"`
}

type Message struct {
	ID    string `json:"id"`
	Time  string `json:"time"`
	From  string `json:"from"`
	Title string `json:"title"`
	Tag   string `json:"tag"`
}

// DashboardService 大屏数据模拟源：内存持有一份基准数据，
// 每次 Snapshot 在上一帧基础上做小幅随机抖动，模拟实时变化。
// 后续接入真实数据时替换本服务实现、保持接口结构即可。
type DashboardService struct {
	mu     sync.Mutex
	cur    Dashboard
	alarms []Alarm
	msgs   []Message
}

func NewDashboardService() *DashboardService {
	trend := make([]TrendPoint, 12)
	for i := range trend {
		trend[i] = TrendPoint{
			T:         twoDigit(i*2) + ":00",
			Intrusion: 2 + rand.Intn(6),
			Fire:      rand.Intn(4),
			Device:    1 + rand.Intn(7),
		}
	}
	return &DashboardService{
		cur: Dashboard{
			Weather: Weather{City: "上海", Temp: 24, Desc: "多云", Icon: "⛅"},
			Env:     Env{Temp: 23.4, Humidity: 56, PM25: 38, CO2: 612},
			Devices: Devices{Online: 1284, Offline: 96, Fault: 23},
			Energy: []EnergyItem{
				{Name: "照明", V: 320, Color: "#0a93c9"},
				{Name: "空调", V: 540, Color: "#2b7fff"},
				{Name: "动力", V: 410, Color: "#1fa85f"},
				{Name: "电梯", V: 180, Color: "#d99a16"},
				{Name: "其他", V: 130, Color: "#ef8326"},
			},
			Parking:   Parking{Total: 1200, Used: 836, Free: 364},
			Workorder: WorkorderSummary{Repair: 42, Inspect: 68, Done: 95, Undone: 15},
			Room:      MachineRoom{Temp: 22.1, Humidity: 45, Power: "A/B 双路", Status: "正常"},
			Broadcast: Broadcast{Zones: 24, Playing: 6, Volume: 72},
			AlarmTrend: trend,
			SecurityCat: []SecurityCat{
				{Name: "视频监控", V: 14, Color: "#0a93c9"},
				{Name: "门禁报警", V: 9, Color: "#2b7fff"},
				{Name: "消防报警", V: 5, Color: "#ef8326"},
				{Name: "周界入侵", V: 7, Color: "#e24550"},
			},
		},
		alarms: []Alarm{
			{ID: "A2041", Time: "14:32:10", Dev: "3F 空调机组 AC-03", Type: "设备故障", Level: "紧急", Desc: "送风温度异常偏高，疑似压缩机过载"},
			{ID: "A2040", Time: "14:18:46", Dev: "B2 配电室 PD-01", Type: "电力报警", Level: "重要", Desc: "A 路电压波动超出阈值 ±10%"},
			{ID: "A2039", Time: "13:55:02", Dev: "周界东侧 PIR-12", Type: "周界入侵", Level: "紧急", Desc: "检测到非法越界，请确认现场情况"},
			{ID: "A2038", Time: "13:40:21", Dev: "5F 烟感 SD-208", Type: "消防报警", Level: "紧急", Desc: "烟雾浓度超标，已联动排烟系统"},
			{ID: "A2037", Time: "13:12:09", Dev: "1F 大堂门禁 ACS-01", Type: "门禁报警", Level: "一般", Desc: "非授权卡片连续刷卡 3 次"},
			{ID: "A2036", Time: "12:48:33", Dev: "机房精密空调 CRAC-2", Type: "环境报警", Level: "重要", Desc: "机房温度 28.6℃ 超过设定上限"},
			{ID: "A2035", Time: "12:20:57", Dev: "P1 车库 CO 探测器", Type: "环境报警", Level: "一般", Desc: "CO 浓度上升，已启动排风"},
		},
		msgs: []Message{
			{ID: "M88", Time: "14:40", From: "运维班组", Title: "今日设备巡检任务已下发", Tag: "任务"},
			{ID: "M87", Time: "14:05", From: "能源管理", Title: "本月累计能耗较上月下降 6.2%", Tag: "报表"},
			{ID: "M86", Time: "13:30", From: "系统", Title: "BA 系统固件升级将于今晚 23:00 进行", Tag: "系统"},
			{ID: "M85", Time: "11:50", From: "物业前台", Title: "3F 会议室预约空调提前开启", Tag: "请求"},
			{ID: "M84", Time: "10:18", From: "安防中心", Title: "周界巡检记录已归档", Tag: "安防"},
			{ID: "M83", Time: "09:02", From: "工单系统", Title: "2 条报修工单待您审批", Tag: "工单"},
		},
	}
}

// Snapshot 返回当前大屏数据，并在内部状态上做一次抖动。
func (s *DashboardService) Snapshot() Dashboard {
	s.mu.Lock()
	defer s.mu.Unlock()

	p := &s.cur
	p.Env = Env{
		Temp:     jitterF(p.Env.Temp, 0.8, 18, 30),
		Humidity: jitterI(p.Env.Humidity, 4, 30, 80),
		PM25:     jitterI(p.Env.PM25, 6, 10, 120),
		CO2:      jitterI(p.Env.CO2, 40, 400, 1200),
	}
	p.Devices = Devices{
		Online:  clampI(p.Devices.Online+rand.Intn(7)-3, 1240, 1320),
		Offline: int(jitterI(float64(p.Devices.Offline), 6, 60, 130)),
		Fault:   int(jitterI(float64(p.Devices.Fault), 3, 8, 40)),
	}
	energy := make([]EnergyItem, len(p.Energy))
	for i, e := range p.Energy {
		e.V = jitterI(e.V, 24, 60, 700)
		energy[i] = e
	}
	p.Energy = energy
	used := int(jitterI(float64(p.Parking.Used), 14, 600, float64(p.Parking.Total)))
	p.Parking = Parking{Total: p.Parking.Total, Used: used, Free: p.Parking.Total - used}
	p.Workorder = WorkorderSummary{
		Repair:  int(jitterI(float64(p.Workorder.Repair), 2, 30, 60)),
		Inspect: int(jitterI(float64(p.Workorder.Inspect), 2, 50, 90)),
		Done:    int(jitterI(float64(p.Workorder.Done), 2, 80, 120)),
		Undone:  int(jitterI(float64(p.Workorder.Undone), 2, 5, 30)),
	}
	p.Room.Temp = jitterF(p.Room.Temp, 0.5, 18, 28)
	p.Room.Humidity = jitterI(p.Room.Humidity, 3, 35, 60)
	p.Broadcast.Playing = int(jitterI(float64(p.Broadcast.Playing), 2, 0, 12))
	p.Broadcast.Volume = int(jitterI(float64(p.Broadcast.Volume), 4, 40, 90))
	trend := make([]TrendPoint, len(p.AlarmTrend))
	for i, x := range p.AlarmTrend {
		x.Intrusion = int(jitterI(float64(x.Intrusion), 2, 0, 12))
		x.Device = int(jitterI(float64(x.Device), 2, 0, 12))
		trend[i] = x
	}
	p.AlarmTrend = trend
	cats := make([]SecurityCat, len(p.SecurityCat))
	for i, sc := range p.SecurityCat {
		sc.V = int(jitterI(float64(sc.V), 2, 0, 30))
		cats[i] = sc
	}
	p.SecurityCat = cats

	return s.cur
}

// Alarms 返回当前未处理的报警列表。
func (s *DashboardService) Alarms() []Alarm {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Alarm, len(s.alarms))
	copy(out, s.alarms)
	return out
}

// HandleAlarm 处理一条报警（从列表移除），返回是否存在。
func (s *DashboardService) HandleAlarm(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, a := range s.alarms {
		if a.ID == id {
			s.alarms = append(s.alarms[:i], s.alarms[i+1:]...)
			return true
		}
	}
	return false
}

// Messages 返回消息列表。
func (s *DashboardService) Messages() []Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Message, len(s.msgs))
	copy(out, s.msgs)
	return out
}

// jitterF 在 ±amp 内随机抖动并夹到 [min,max]，保留 1 位小数。
func jitterF(v, amp, min, max float64) float64 {
	v += (rand.Float64() - 0.5) * 2 * amp
	return math.Round(math.Min(math.Max(v, min), max)*10) / 10
}

// jitterI 同 jitterF，但取整。
func jitterI(v, amp, min, max float64) float64 {
	v += (rand.Float64() - 0.5) * 2 * amp
	return math.Round(math.Min(math.Max(v, min), max))
}

func clampI(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func twoDigit(n int) string {
	return fmt.Sprintf("%02d", n)
}
