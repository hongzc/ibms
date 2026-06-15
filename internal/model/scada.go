package model

import (
	"encoding/json"
	"time"
)

// ScadaProject 组态工程：name 为工程名，Graph 为 X6 画布的 JSON 序列化（原样透传）。
type ScadaProject struct {
	ID        int64           `json:"id"`
	Name      string          `json:"name"`
	Graph     json.RawMessage `json:"graph,omitempty"`
	UpdatedAt time.Time       `json:"updatedAt"`
}
