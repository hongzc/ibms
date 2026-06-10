package route

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"private/ibms/internal/service"
)

type dashboardHandler struct {
	svc *service.DashboardService
}

func (h *dashboardHandler) snapshot(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.Snapshot())
}

func (h *dashboardHandler) alarms(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.Alarms())
}

type handleAlarmRequest struct {
	ID string `json:"id" binding:"required"`
}

func (h *dashboardHandler) handleAlarm(c *gin.Context) {
	var req handleAlarmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !h.svc.HandleAlarm(req.ID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "alarm not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *dashboardHandler) messages(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.Messages())
}
