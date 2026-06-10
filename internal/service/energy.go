package service

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

type EnergyTotals struct {
	Today     float64 `json:"today"`
	Predict   float64 `json:"predict"`
	Yesterday float64 `json:"yesterday"`
	Month     float64 `json:"month"`
	LastMonth float64 `json:"lastMonth"`
	Year      float64 `json:"year"`
	Unit      string  `json:"unit"`
}

type EnergyTrendPoint struct {
	T     string  `json:"t"`
	Elec  float64 `json:"elec"`
	Water float64 `json:"water"`
}

type EnergyOverview struct {
	Elec    EnergyTotals       `json:"elec"`
	Water   EnergyTotals       `json:"water"`
	Trend   []EnergyTrendPoint `json:"trend24h"`
	Rank    []NameValue        `json:"rank"`
}

type NameValue struct {
	Name string  `json:"name"`
	V    float64 `json:"v"`
}

type EnergyCompare struct {
	Name     string  `json:"name"`
	Current  float64 `json:"current"`
	Previous float64 `json:"previous"`
	YoY      float64 `json:"yoy"`
	MoM      float64 `json:"mom"`
}

type EnergyRecent struct {
	T        string  `json:"t"`
	Current  float64 `json:"current"`
	Previous float64 `json:"previous"`
}

type EnergyAnalysis struct {
	Compare []EnergyCompare `json:"compare"`
	Share   []NameValue     `json:"share"`
	Recent  []EnergyRecent  `json:"recent"`
}

type EnergyMeter struct {
	ID      string      `json:"id"`
	Name    string      `json:"name"`
	Kind    string      `json:"kind"` // 电表 | 水表
	Zone    string      `json:"zone"`
	Status  string      `json:"status"`
	Params  MeterParams `json:"params"`
	Reading float64     `json:"reading"`
}

type MeterParams struct {
	Voltage *float64 `json:"voltage"`
	Current *float64 `json:"current"`
	Power   *float64 `json:"power"`
	Flow    *float64 `json:"flow"`
}

type MeterReading struct {
	Meter string  `json:"meter"`
	Kind  string  `json:"kind"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Usage float64 `json:"usage"`
	Date  string  `json:"date"`
}

// EnergyService 能源数据模拟源。
type EnergyService struct {
	mu     sync.Mutex
	meters []EnergyMeter
}

var energyDims = map[string][]string{
	"item":     {"照明", "空调", "动力", "电梯", "其他"},
	"building": {"A 栋", "B 栋", "C 栋", "地下室"},
	"function": {"办公", "商业", "餐饮", "车库", "公区"},
}

func NewEnergyService() *EnergyService {
	s := &EnergyService{}
	f := func(v float64) *float64 { return &v }
	zones := []string{"A栋 3F", "A栋 8F", "B栋 2F", "B栋 9F", "C栋 1F", "地下室 B1"}
	for i := 0; i < 10; i++ {
		s.meters = append(s.meters, EnergyMeter{
			ID: fmt.Sprintf("EM-%02d", i+1), Name: fmt.Sprintf("%s %s电表", zones[i%len(zones)], []string{"照明", "空调", "动力"}[i%3]),
			Kind: "电表", Zone: zones[i%len(zones)], Status: "正常",
			Params:  MeterParams{Voltage: f(220 + rand.Float64()*12), Current: f(10 + rand.Float64()*30), Power: f(2 + rand.Float64()*8)},
			Reading: 30000 + rand.Float64()*30000,
		})
	}
	for i := 0; i < 4; i++ {
		s.meters = append(s.meters, EnergyMeter{
			ID: fmt.Sprintf("WM-%02d", i+1), Name: fmt.Sprintf("%s 水表", zones[i%len(zones)]),
			Kind: "水表", Zone: zones[i%len(zones)], Status: "正常",
			Params:  MeterParams{Flow: f(0.5 + rand.Float64()*3)},
			Reading: 4000 + rand.Float64()*5000,
		})
	}
	s.meters[7].Status = "离线"
	return s
}

// Overview 系统态势：水电 24h 趋势 + 各周期用量。
func (s *EnergyService) Overview() EnergyOverview {
	now := time.Now()
	hour := now.Hour()
	trend := make([]EnergyTrendPoint, 24)
	var todayElec, todayWater float64
	for i := 0; i < 24; i++ {
		// 营业时间负荷高的日负荷曲线
		base := 60 + 120*math.Exp(-math.Pow(float64(i)-14, 2)/40)
		e := jitterF(base, 12, 30, 260)
		w := jitterF(base/18, 1.2, 1, 16)
		if i > hour { // 未来时段无数据
			e, w = 0, 0
		} else {
			todayElec += e
			todayWater += w
		}
		trend[i] = EnergyTrendPoint{T: fmt.Sprintf("%02d:00", i), Elec: e, Water: w}
	}
	progress := math.Max(float64(hour)/24, 0.05)
	return EnergyOverview{
		Elec: EnergyTotals{
			Today: math.Round(todayElec), Predict: math.Round(todayElec / progress),
			Yesterday: 5980, Month: 128600, LastMonth: 135400, Year: 1486000, Unit: "kWh",
		},
		Water: EnergyTotals{
			Today: math.Round(todayWater), Predict: math.Round(todayWater / progress),
			Yesterday: 438, Month: 9620, LastMonth: 10100, Year: 112000, Unit: "m³",
		},
		Trend: trend,
		Rank: []NameValue{
			{"空调", jitterI(540, 20, 480, 600)}, {"动力", jitterI(410, 16, 360, 460)},
			{"照明", jitterI(320, 14, 280, 360)}, {"电梯", jitterI(180, 10, 150, 210)},
			{"其他", jitterI(130, 8, 100, 160)},
		},
	}
}

// Analysis 能耗分析：按维度与周期返回同环比、占比、近若干天对比。
func (s *EnergyService) Analysis(dim, period string) EnergyAnalysis {
	names, ok := energyDims[dim]
	if !ok {
		names = energyDims["item"]
	}
	scale := map[string]float64{"day": 1, "month": 28, "year": 330}[period]
	if scale == 0 {
		scale = 1
	}
	var out EnergyAnalysis
	for i, n := range names {
		base := float64(600-i*100) * scale
		cur := jitterF(base, base*0.08, base*0.7, base*1.3)
		prev := jitterF(base, base*0.1, base*0.7, base*1.3)
		out.Compare = append(out.Compare, EnergyCompare{
			Name: n, Current: math.Round(cur), Previous: math.Round(prev),
			YoY: math.Round((cur-prev)/prev*1000) / 10,
			MoM: math.Round((rand.Float64()-0.45)*160) / 10,
		})
		out.Share = append(out.Share, NameValue{Name: n, V: math.Round(cur)})
	}
	days := 7
	if period != "day" {
		days = 30
	}
	for i := days - 1; i >= 0; i-- {
		d := time.Now().AddDate(0, 0, -i)
		out.Recent = append(out.Recent, EnergyRecent{
			T:        d.Format("01-02"),
			Current:  jitterI(5400, 500, 4200, 6600),
			Previous: jitterI(5200, 500, 4000, 6400),
		})
	}
	return out
}

// Meters 仪表实时参数（带抖动）。
func (s *EnergyService) Meters() []EnergyMeter {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]EnergyMeter, len(s.meters))
	for i := range s.meters {
		m := &s.meters[i]
		if m.Status == "正常" {
			if m.Params.Voltage != nil {
				*m.Params.Voltage = jitterF(*m.Params.Voltage, 1.5, 215, 235)
				*m.Params.Current = jitterF(*m.Params.Current, 1.2, 5, 45)
				*m.Params.Power = jitterF(*m.Params.Power, 0.4, 1, 12)
				m.Reading = math.Round((m.Reading+*m.Params.Power/12)*10) / 10
			}
			if m.Params.Flow != nil {
				*m.Params.Flow = jitterF(*m.Params.Flow, 0.3, 0.2, 5)
				m.Reading = math.Round((m.Reading+*m.Params.Flow/30)*10) / 10
			}
		}
		out[i] = *m
	}
	return out
}

// Readings 近 N 天抄表记录。
func (s *EnergyService) Readings(days int) []MeterReading {
	if days <= 0 || days > 31 {
		days = 7
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []MeterReading
	for d := 0; d < days; d++ {
		date := time.Now().AddDate(0, 0, -d-1).Format("2006-01-02")
		for _, m := range s.meters {
			usage := jitterF(230, 40, 120, 360)
			if m.Kind == "水表" {
				usage = jitterF(14, 4, 4, 30)
			}
			end := m.Reading - float64(d)*usage
			out = append(out, MeterReading{
				Meter: m.ID, Kind: m.Kind,
				Start: math.Round((end-usage)*10) / 10, End: math.Round(end*10) / 10,
				Usage: math.Round(usage*10) / 10, Date: date,
			})
		}
	}
	return out
}
