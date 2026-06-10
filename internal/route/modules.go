package route

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"private/ibms/internal/service"
)

/* ---------------- BA 设备监控 ---------------- */

type baHandler struct {
	svc *service.BAService
}

// baType 校验并返回设备类型参数，非法时回 400 并返回 false。
func baType(c *gin.Context) (string, bool) {
	typ := c.DefaultQuery("type", "fresh")
	if typ != "fresh" && typ != "ac" && typ != "chiller" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid type"})
		return "", false
	}
	return typ, true
}

func (h *baHandler) devices(c *gin.Context) {
	typ, ok := baType(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, h.svc.Devices(typ))
}

type baControlRequest struct {
	ID     string `json:"id"`
	Action string `json:"action" binding:"required"`
	Value  any    `json:"value"`
}

func (h *baHandler) control(c *gin.Context) {
	var req baControlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}
	d, err := h.svc.Control(req.ID, req.Action, req.Value)
	if err != nil {
		if errors.Is(err, service.ErrBADeviceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, d)
}

func (h *baHandler) controlAll(c *gin.Context) {
	typ, ok := baType(c)
	if !ok {
		return
	}
	var req baControlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	n, err := h.svc.ControlAll(typ, req.Action)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"affected": n})
}

func (h *baHandler) events(c *gin.Context) {
	typ, ok := baType(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, h.svc.Events(typ))
}

/* ---------------- 能源管理 ---------------- */

type energyHandler struct {
	svc *service.EnergyService
}

func (h *energyHandler) overview(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.Overview())
}

func (h *energyHandler) analysis(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.Analysis(c.DefaultQuery("dim", "item"), c.DefaultQuery("period", "day")))
}

func (h *energyHandler) meters(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.Meters())
}

func (h *energyHandler) readings(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	c.JSON(http.StatusOK, h.svc.Readings(days))
}

/* ---------------- 运营管理 ---------------- */

type operationHandler struct {
	svc *service.OperationService
}

func (h *operationHandler) overview(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.Overview())
}

func (h *operationHandler) rooms(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.Rooms(c.Query("building")))
}

func (h *operationHandler) leases(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.Leases())
}

/* ---------------- 工单系统 ---------------- */

type workorderHandler struct {
	svc *service.WorkorderService
}

func (h *workorderHandler) board(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.Board())
}

type createWorkorderRequest struct {
	Title    string `json:"title" binding:"required"`
	Type     string `json:"type" binding:"required"`
	Location string `json:"location"`
	Desc     string `json:"desc"`
}

func (h *workorderHandler) create(c *gin.Context) {
	var req createWorkorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, h.svc.Create(req.Title, req.Type, req.Location, req.Desc))
}

type dispatchRequest struct {
	ID       string `json:"id" binding:"required"`
	Assignee string `json:"assignee" binding:"required"`
}

func (h *workorderHandler) dispatch(c *gin.Context) {
	var req dispatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	w, err := h.svc.Dispatch(req.ID, req.Assignee)
	if err != nil {
		workorderError(c, err)
		return
	}
	c.JSON(http.StatusOK, w)
}

type completeRequest struct {
	ID string `json:"id" binding:"required"`
}

func (h *workorderHandler) complete(c *gin.Context) {
	var req completeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	w, err := h.svc.Complete(req.ID)
	if err != nil {
		workorderError(c, err)
		return
	}
	c.JSON(http.StatusOK, w)
}

func workorderError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrWorkorderNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}
