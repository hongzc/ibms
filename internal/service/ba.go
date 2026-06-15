package service

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// BADevice 楼宇自控设备（新风/空调/冷热源通用结构）。
type BADevice struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Type   string   `json:"type"` // fresh | ac | chiller
	Zone   string   `json:"zone"`
	Online bool     `json:"online"`
	Status string   `json:"status"` // 运行 | 停止 | 故障
	Mode   string   `json:"mode"`   // 自动 | 手动
	Params BAParams `json:"params"`
}

type BAParams struct {
	SupplyTemp float64 `json:"supplyTemp"`
	ReturnTemp float64 `json:"returnTemp"`
	Humidity   float64 `json:"humidity"`
	Valve      float64 `json:"valve"`
	FanFreq    float64 `json:"fanFreq"`
	Power      float64 `json:"power"`
	SetTemp    float64 `json:"setTemp"`
}

type BAStats struct {
	Total   int `json:"total"`
	Online  int `json:"online"`
	Running int `json:"running"`
	Stopped int `json:"stopped"`
	Fault   int `json:"fault"`
}

type BAZone struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type BADeviceList struct {
	Stats   BAStats    `json:"stats"`
	Zones   []BAZone   `json:"zones"`
	Devices []BADevice `json:"devices"`
}

type BAEvent struct {
	Time   string `json:"time"`
	Device string `json:"device"`
	Event  string `json:"event"`
	By     string `json:"by"`
	Result string `json:"result"`
}

// BAService BA 设备模拟源：内存设备台账 + 运行参数抖动 + 控制回写。
type BAService struct {
	mu      sync.Mutex
	devices []BADevice
	events  map[string][]BAEvent // 按设备类型分组
}

func NewBAService() *BAService {
	s := &BAService{events: map[string][]BAEvent{}}
	mk := func(id, name, typ, zone string, running bool) BADevice {
		status := "运行"
		if !running {
			status = "停止"
		}
		return BADevice{
			ID: id, Name: name, Type: typ, Zone: zone, Online: true, Status: status, Mode: "自动",
			Params: BAParams{
				SupplyTemp: round1(17 + rand.Float64()*4), ReturnTemp: round1(22 + rand.Float64()*3),
				Humidity: round1(45 + rand.Float64()*20), Valve: round1(30 + rand.Float64()*50),
				FanFreq: round1(30 + rand.Float64()*15), Power: round1(8 + rand.Float64()*20),
				SetTemp: 24,
			},
		}
	}
	zonesA := []string{"A 栋 1F", "A 栋 3F", "A 栋 8F", "A 栋 12F"}
	zonesB := []string{"B 栋 2F", "B 栋 5F", "B 栋 9F"}
	for i := 1; i <= 8; i++ {
		s.devices = append(s.devices, mk(fmt.Sprintf("AHU-A-%02d", i), fmt.Sprintf("新风机组 AHU-A-%02d", i), "fresh", zonesA[(i-1)%len(zonesA)], i%5 != 0))
	}
	for i := 1; i <= 6; i++ {
		s.devices = append(s.devices, mk(fmt.Sprintf("AHU-B-%02d", i), fmt.Sprintf("新风机组 AHU-B-%02d", i), "fresh", zonesB[(i-1)%len(zonesB)], i%4 != 0))
	}
	for i := 1; i <= 10; i++ {
		s.devices = append(s.devices, mk(fmt.Sprintf("AC-A-%02d", i), fmt.Sprintf("空调机组 AC-A-%02d", i), "ac", zonesA[(i-1)%len(zonesA)], i%6 != 0))
	}
	for i := 1; i <= 8; i++ {
		s.devices = append(s.devices, mk(fmt.Sprintf("AC-B-%02d", i), fmt.Sprintf("空调机组 AC-B-%02d", i), "ac", zonesB[(i-1)%len(zonesB)], true))
	}
	chZones := []string{"地下室 B1 冷冻机房", "地下室 B2 锅炉房"}
	for i := 1; i <= 6; i++ {
		s.devices = append(s.devices, mk(fmt.Sprintf("CH-%02d", i), fmt.Sprintf("冷热源机组 CH-%02d", i), "chiller", chZones[(i-1)%2], i%3 != 0))
	}
	// 随机标一两台故障
	s.devices[4].Status = "故障"
	s.devices[4].Online = false
	s.devices[20].Status = "故障"

	for _, typ := range []string{"fresh", "ac", "chiller"} {
		s.events[typ] = []BAEvent{
			{Time: "13:40:21", Device: s.firstOf(typ), Event: "启动", By: "自动策略", Result: "成功"},
			{Time: "12:18:09", Device: s.firstOf(typ), Event: "温度设定 24.0℃", By: "管理员", Result: "成功"},
			{Time: "09:02:44", Device: s.firstOf(typ), Event: "切换自动模式", By: "管理员", Result: "成功"},
		}
	}
	return s
}

func (s *BAService) firstOf(typ string) string {
	for _, d := range s.devices {
		if d.Type == typ {
			return d.ID
		}
	}
	return ""
}

// Devices 按类型返回设备列表与统计，运行参数做小幅抖动。
func (s *BAService) Devices(typ string) BADeviceList {
	s.mu.Lock()
	defer s.mu.Unlock()

	var out BADeviceList
	zoneCount := map[string]int{}
	var zoneOrder []string
	for i := range s.devices {
		d := &s.devices[i]
		if d.Type != typ {
			continue
		}
		if d.Status == "运行" {
			d.Params.SupplyTemp = jitterF(d.Params.SupplyTemp, 0.4, 15, 24)
			d.Params.ReturnTemp = jitterF(d.Params.ReturnTemp, 0.4, 20, 28)
			d.Params.Humidity = jitterI(d.Params.Humidity, 3, 35, 75)
			d.Params.FanFreq = jitterF(d.Params.FanFreq, 1.5, 25, 50)
			d.Params.Power = jitterF(d.Params.Power, 1.2, 5, 35)
		}
		out.Devices = append(out.Devices, *d)
		out.Stats.Total++
		if d.Online {
			out.Stats.Online++
		}
		switch d.Status {
		case "运行":
			out.Stats.Running++
		case "停止":
			out.Stats.Stopped++
		case "故障":
			out.Stats.Fault++
		}
		zone := zoneBuilding(d.Zone)
		if _, ok := zoneCount[zone]; !ok {
			zoneOrder = append(zoneOrder, zone)
		}
		zoneCount[zone]++
	}
	for _, z := range zoneOrder {
		out.Zones = append(out.Zones, BAZone{Name: z, Count: zoneCount[z]})
	}
	return out
}

// zoneBuilding 取区域的楼栋前缀（"A 栋 3F" → "A 栋"）。
func zoneBuilding(zone string) string {
	runes := []rune(zone)
	for i, r := range runes {
		if r == '栋' {
			return string(runes[:i+1])
		}
	}
	return zone
}

var ErrBADeviceNotFound = fmt.Errorf("device not found")
var ErrBABadAction = fmt.Errorf("invalid action")

// Control 对单台设备下发控制，返回更新后的设备。
func (s *BAService) Control(id, action string, value any) (BADevice, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.devices {
		d := &s.devices[i]
		if d.ID != id {
			continue
		}
		desc := ""
		switch action {
		case "start":
			d.Status = "运行"
			desc = "启动"
		case "stop":
			d.Status = "停止"
			desc = "停止"
		case "mode":
			m, _ := value.(string)
			if m != "自动" && m != "手动" {
				return BADevice{}, ErrBABadAction
			}
			d.Mode = m
			desc = "切换" + m + "模式"
		case "setTemp":
			v, ok := toFloat(value)
			if !ok {
				return BADevice{}, ErrBABadAction
			}
			d.Params.SetTemp = round1(v)
			desc = fmt.Sprintf("温度设定 %.1f℃", v)
		case "setValve":
			v, ok := toFloat(value)
			if !ok {
				return BADevice{}, ErrBABadAction
			}
			d.Params.Valve = round1(v)
			desc = fmt.Sprintf("阀门开度 %.0f%%", v)
		default:
			return BADevice{}, ErrBABadAction
		}
		s.appendEvent(d.Type, BAEvent{
			Time: time.Now().Format("15:04:05"), Device: d.ID, Event: desc, By: "管理员", Result: "成功",
		})
		return *d, nil
	}
	return BADevice{}, ErrBADeviceNotFound
}

// ControlAll 按类型群控启停，返回影响台数。
func (s *BAService) ControlAll(typ, action string) (int, error) {
	if action != "start" && action != "stop" {
		return 0, ErrBABadAction
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for i := range s.devices {
		d := &s.devices[i]
		if d.Type != typ || d.Status == "故障" {
			continue
		}
		if action == "start" {
			d.Status = "运行"
		} else {
			d.Status = "停止"
		}
		n++
	}
	desc := "群控启动"
	if action == "stop" {
		desc = "群控停止"
	}
	s.appendEvent(typ, BAEvent{
		Time: time.Now().Format("15:04:05"), Device: "全部设备", Event: fmt.Sprintf("%s（%d 台）", desc, n), By: "管理员", Result: "成功",
	})
	return n, nil
}

// List 返回全部设备快照（不抖动），供组态点表枚举。
func (s *BAService) List() []BADevice {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]BADevice, len(s.devices))
	copy(out, s.devices)
	return out
}

// Events 返回某类型设备的运行事件（新事件在前）。
func (s *BAService) Events(typ string) []BAEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]BAEvent, len(s.events[typ]))
	copy(out, s.events[typ])
	return out
}

func (s *BAService) appendEvent(typ string, e BAEvent) {
	evs := append([]BAEvent{e}, s.events[typ]...)
	if len(evs) > 30 {
		evs = evs[:30]
	}
	s.events[typ] = evs
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	}
	return 0, false
}
