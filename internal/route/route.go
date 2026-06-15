package route

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"private/ibms/internal/service"
	"private/ibms/web"
)

// New 构建 gin 引擎并注册路由。
func New(svc *service.Service) *gin.Engine {
	r := gin.Default()

	// 静态页面：综合态势 + 各子模块大屏
	page := func(name string) gin.HandlerFunc {
		return func(c *gin.Context) {
			data, err := web.FS.ReadFile(name)
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			c.Data(http.StatusOK, "text/html; charset=utf-8", data)
		}
	}
	r.GET("/", page("index.html"))
	r.GET("/ba", page("ba.html"))
	r.GET("/energy", page("energy.html"))
	r.GET("/operation", page("operation.html"))
	r.GET("/workorder", page("workorder.html"))
	r.GET("/editor", page("editor.html"))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	h := &userHandler{svc: svc.User}
	d := &dashboardHandler{svc: svc.Dashboard}
	ba := &baHandler{svc: svc.BA}
	en := &energyHandler{svc: svc.Energy}
	op := &operationHandler{svc: svc.Operation}
	wo := &workorderHandler{svc: svc.Workorder}
	sc := &scadaHandler{svc: svc.Scada}
	api := r.Group("/api/v1")
	{
		api.POST("/users", h.create)
		api.GET("/users/detail", h.get)
		api.GET("/users", h.list)

		api.GET("/dashboard", d.snapshot)
		api.GET("/alarms", d.alarms)
		api.POST("/alarms/handle", d.handleAlarm)
		api.GET("/messages", d.messages)

		api.GET("/ba/devices", ba.devices)
		api.POST("/ba/devices/control-all", ba.controlAll)
		api.POST("/ba/devices/control", ba.control)
		api.GET("/ba/events", ba.events)

		api.GET("/energy/overview", en.overview)
		api.GET("/energy/analysis", en.analysis)
		api.GET("/energy/meters", en.meters)
		api.GET("/energy/readings", en.readings)

		api.GET("/operation/overview", op.overview)
		api.GET("/operation/rooms", op.rooms)
		api.GET("/operation/leases", op.leases)

		api.GET("/workorders", wo.board)
		api.POST("/workorders", wo.create)
		api.POST("/workorders/dispatch", wo.dispatch)
		api.POST("/workorders/complete", wo.complete)

		api.GET("/scada/projects", sc.list)
		api.GET("/scada/projects/detail", sc.detail)
		api.POST("/scada/projects", sc.create)
		api.POST("/scada/projects/save", sc.save)
		api.POST("/scada/projects/delete", sc.delete)
		api.POST("/scada/publish", sc.publish)
		api.GET("/scada/published", sc.published)
		api.GET("/scada/datapoints", sc.datapoints)
		api.GET("/scada/values", sc.values)
	}

	return r
}
