package route

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"private/ibms/internal/service"
)

// New 构建 gin 引擎并注册路由。
func New(svc *service.Service) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	h := &userHandler{svc: svc.User}
	api := r.Group("/api/v1")
	{
		api.POST("/users", h.create)
		api.GET("/users/:id", h.get)
		api.GET("/users", h.list)
	}

	return r
}
