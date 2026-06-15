package service

import (
	"context"
	"encoding/json"
	"fmt"

	"private/ibms/internal/model"
	"private/ibms/internal/store"
)

// ScadaService 组态工程服务：工程文件 CRUD（落 SQLite）+ 数据点表/实时值（取自 BA 设备）。
type ScadaService struct {
	store store.ScadaStore
	ba    *BAService
}

func NewScadaService(s store.ScadaStore, ba *BAService) *ScadaService {
	return &ScadaService{store: s, ba: ba}
}

func (s *ScadaService) Create(ctx context.Context, name string) (*model.ScadaProject, error) {
	id, err := s.store.Create(ctx, name, "")
	if err != nil {
		return nil, err
	}
	return s.store.Get(ctx, id)
}

func (s *ScadaService) Save(ctx context.Context, id int64, name, graph string) error {
	return s.store.Save(ctx, id, name, graph)
}

func (s *ScadaService) Get(ctx context.Context, id int64) (*model.ScadaProject, error) {
	return s.store.Get(ctx, id)
}

func (s *ScadaService) List(ctx context.Context) ([]*model.ScadaProject, error) {
	return s.store.List(ctx)
}

func (s *ScadaService) Delete(ctx context.Context, id int64) error {
	return s.store.Delete(ctx, id)
}

// Publish 中台「下发到大屏」：标记该工程为大屏二维组态视图当前展示的图纸。
func (s *ScadaService) Publish(ctx context.Context, id int64) error {
	return s.store.SetPublished(ctx, id)
}

// Published 返回当前下发给大屏的工程（含 graph）；未显式下发时回退为最近保存的一个。
func (s *ScadaService) Published(ctx context.Context) (*model.ScadaProject, error) {
	id, err := s.store.GetPublishedID(ctx)
	if err != nil {
		return nil, err
	}
	if id == 0 {
		list, err := s.store.List(ctx)
		if err != nil || len(list) == 0 {
			return nil, err
		}
		id = list[0].ID // List 按 updated_at DESC，取最近
	}
	return s.store.Get(ctx, id)
}

// DataPoint 可绑定数据点（点表项）。
type DataPoint struct {
	ID     string `json:"id"`     // "AHU-A-01.status"
	Label  string `json:"label"`  // "新风机组 AHU-A-01 · 运行状态"
	Device string `json:"device"` // "AHU-A-01"
	Field  string `json:"field"`  // "status"
	Kind   string `json:"kind"`   // status | number
	Unit   string `json:"unit"`
}

// baPointFields 每台 BA 设备暴露的可绑定字段。
var baPointFields = []struct{ field, label, kind, unit string }{
	{"status", "运行状态", "status", ""},
	{"params.supplyTemp", "送风/出水温度", "number", "℃"},
	{"params.returnTemp", "回风/回水温度", "number", "℃"},
	{"params.power", "运行功率", "number", "kW"},
	{"params.valve", "阀门开度", "number", "%"},
	{"params.fanFreq", "风机频率", "number", "Hz"},
}

// DataPoints 返回全部可绑定数据点（按设备展开字段）。
func (s *ScadaService) DataPoints() []DataPoint {
	var out []DataPoint
	for _, d := range s.ba.List() {
		for _, f := range baPointFields {
			out = append(out, DataPoint{
				ID:     d.ID + "." + f.field,
				Label:  d.Name + " · " + f.label,
				Device: d.ID,
				Field:  f.field,
				Kind:   f.kind,
				Unit:   f.unit,
			})
		}
	}
	return out
}

// Values 返回全部数据点的当前值映射（带抖动，供运行态实时刷新）。
func (s *ScadaService) Values() map[string]any {
	out := map[string]any{}
	for _, typ := range []string{"fresh", "ac", "chiller"} {
		for _, d := range s.ba.Devices(typ).Devices {
			for _, f := range baPointFields {
				out[d.ID+"."+f.field] = baFieldValue(d, f.field)
			}
		}
	}
	return out
}

// baFieldValue 按字段路径取 BA 设备的值。
func baFieldValue(d BADevice, field string) any {
	switch field {
	case "status":
		return d.Status
	case "params.supplyTemp":
		return d.Params.SupplyTemp
	case "params.returnTemp":
		return d.Params.ReturnTemp
	case "params.power":
		return d.Params.Power
	case "params.valve":
		return d.Params.Valve
	case "params.fanFreq":
		return d.Params.FanFreq
	}
	return nil
}

// GraphString 把工程的 Graph 原样取出为字符串（空则返回 "{}"）。
func GraphString(p *model.ScadaProject) string {
	if p == nil || len(p.Graph) == 0 {
		return "{}"
	}
	return string(p.Graph)
}

// MarshalGraph 校验前端传来的 graph 是合法 JSON，返回紧凑字符串。
func MarshalGraph(raw json.RawMessage) (string, error) {
	if len(raw) == 0 {
		return "{}", nil
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return "", fmt.Errorf("graph 不是合法 JSON: %w", err)
	}
	return string(raw), nil
}
