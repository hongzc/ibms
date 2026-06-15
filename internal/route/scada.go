package route

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"private/ibms/internal/service"
)

type scadaHandler struct {
	svc *service.ScadaService
}

func (h *scadaHandler) list(c *gin.Context) {
	ps, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ps)
}

func (h *scadaHandler) detail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Query("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	p, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if p == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	c.JSON(http.StatusOK, p)
}

type createScadaRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *scadaHandler) create(c *gin.Context) {
	var req createScadaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.svc.Create(c.Request.Context(), req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

type saveScadaRequest struct {
	ID    int64           `json:"id" binding:"required"`
	Name  string          `json:"name" binding:"required"`
	Graph json.RawMessage `json:"graph"`
}

func (h *scadaHandler) save(c *gin.Context) {
	var req saveScadaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	graph, err := service.MarshalGraph(req.Graph)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Save(c.Request.Context(), req.ID, req.Name, graph); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type deleteScadaRequest struct {
	ID int64 `json:"id" binding:"required"`
}

func (h *scadaHandler) delete(c *gin.Context) {
	var req deleteScadaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Delete(c.Request.Context(), req.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type publishScadaRequest struct {
	ID int64 `json:"id" binding:"required"`
}

func (h *scadaHandler) publish(c *gin.Context) {
	var req publishScadaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Publish(c.Request.Context(), req.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// published 大屏二维组态视图拉取「已下发」的组态图纸（只读展示用）。
func (h *scadaHandler) published(c *gin.Context) {
	p, err := h.svc.Published(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if p == nil {
		c.JSON(http.StatusOK, gin.H{}) // 未下发，大屏显示空态
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *scadaHandler) datapoints(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.DataPoints())
}

func (h *scadaHandler) values(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.Values())
}
