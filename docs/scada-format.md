# 组态（二维）数据格式与对接契约

> 给中台/组态编辑器开发方的接口规范。
> **中台自建组态编辑器，技术栈不限**（Vue/React/原生/任意画布库），只要：
> 1. 下发的图纸是本文「组态图纸格式」的 JSON；
> 2. 图元的 `binding` 与「数据点 / 实时值契约」的 id 对齐。
> 大屏（渲染端）即可正确加载并按实时数据展示、控制。
>
> 本仓库 `/editor` 是一个**参考实现（Mock）**，可照着它的行为与本文格式自建生产版本。

---

## 一、组态图纸格式（中立格式，与渲染库无关）

工程的图纸是一个 JSON 对象：

```json
{
  "version": 1,
  "nodes": [
    { "id": "n1", "type": "chiller", "x": 340, "y": 200, "label": "1#冷热源机组", "binding": "CH-01.status", "unit": "" },
    { "id": "n2", "type": "pump",    "x": 380, "y": 360, "label": "冷冻水泵",      "binding": "CH-02.status", "unit": "" },
    { "id": "n3", "type": "vdisp",   "x": 560, "y": 360, "label": "机组功率",      "binding": "CH-01.params.power", "unit": "kW" }
  ],
  "edges": [
    { "from": "n1", "to": "n2" },
    { "from": "n2", "to": "n3" }
  ]
}
```

### node 字段

| 字段 | 必填 | 说明 |
|---|---|---|
| `id` | ✓ | 节点唯一 id（edges 用它引用） |
| `type` | ✓ | 图元类型，见下表 |
| `x` `y` | ✓ | 画布坐标（示意布局，**不代表设备物理位置**） |
| `label` | | 显示名称 |
| `binding` | | 绑定的数据点 id（见第二节）；不绑则留空 |
| `unit` | | 仅 `vdisp` 数显框用，如 `kW`、`℃` |

### edge 字段

| 字段 | 必填 | 说明 |
|---|---|---|
| `from` | ✓ | 起点 node id |
| `to` | ✓ | 终点 node id |
| `fromPort` `toPort` | | 可选，端口 `t`/`r`/`b`/`l`（上右下左）；省略则自动锚点 |

### 图元类型（type）

| type | 含义 | 绑定字段类型 |
|---|---|---|
| `pump` | 水泵 | 状态 |
| `fan` | 风机 | 状态 |
| `valve` | 阀门 | 状态 |
| `chiller` | 冷热源机组 | 状态 |
| `vdisp` | 数显框 | 数值 |
| `indicator` | 指示灯 | 状态 |
| `text` | 文本 | 无 |
| （edge）| 管道 | — |

---

## 二、数据点 / 实时值契约

图元的 `binding` 必须等于实时值接口里的某个 key。

### 数据点 id 约定

`<设备id>.<字段>`，例如 `CH-02.status`、`AHU-A-01.params.power`。
当前可用字段：`status`、`params.supplyTemp`、`params.returnTemp`、`params.power`、`params.valve`、`params.fanFreq`。

### `GET /api/v1/scada/datapoints` — 可绑定点清单

编辑器的「绑定数据点」下拉读它。

```json
[
  { "id": "CH-02.status", "label": "冷热源机组 CH-02 · 运行状态",
    "device": "CH-02", "field": "status", "kind": "status", "unit": "" },
  { "id": "CH-02.params.power", "label": "冷热源机组 CH-02 · 运行功率",
    "device": "CH-02", "field": "params.power", "kind": "number", "unit": "kW" }
]
```

`kind` 取 `status`（状态字符串）或 `number`（数值）。

### `GET /api/v1/scada/values` — 实时值（大屏每 2 秒轮询）

扁平 map：`数据点id → 当前值`。

```json
{ "CH-02.status": "运行", "CH-01.params.power": 12.4 }
```

- 状态字段值为字符串：`运行` / `停止` / `故障`
- 数值字段值为数字

---

## 三、渲染语义（大屏如何根据值展示）

| 图元 | 绑定 | 表现 |
|---|---|---|
| pump / fan | 状态 | `运行`→绿色+旋转；`停止`→灰；`故障`→红 |
| valve / chiller / indicator | 状态 | `运行`→绿；`停止`→灰；`故障`→红 |
| 上述图元 | 数值 | >5 视为运行（绿），否则停止（灰） |
| vdisp 数显框 | 数值 | 显示「数值 + 单位」 |
| 管道（edge） | — | 起点节点为「运行」时显示流动动画 |

---

## 四、下发与展示链路

```
中台组态编辑器 ──下发(图纸JSON)──▶ POST /api/v1/scada/publish
                                          │
                              GET /api/v1/scada/published  ◀── 大屏「二维组态」视图加载
                              GET /api/v1/scada/values      ◀── 大屏每 2 秒轮询刷新
```

### 相关接口一览

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/v1/scada/projects` | 工程列表 |
| GET | `/api/v1/scada/projects/detail?id=` | 取工程（含图纸） |
| POST | `/api/v1/scada/projects` | 新建 `{name}` |
| POST | `/api/v1/scada/projects/save` | 保存 `{id,name,graph}`，`graph` 即第一节 JSON |
| POST | `/api/v1/scada/projects/delete` | 删除 `{id}` |
| POST | `/api/v1/scada/publish` | 下发到大屏 `{id}` |
| GET | `/api/v1/scada/published` | 大屏取当前下发的工程 |
| GET | `/api/v1/scada/datapoints` | 可绑定点清单 |
| GET | `/api/v1/scada/values` | 实时值 map |

> 说明：本仓库后端的数据点 / 实时值取自 BA 设备的 Mock。接真实数据时，中台只需保证 `binding` 的 id 与实时值 map 的 key 一致即可，设备种类不限——**靠 id 字符串对齐**。
