package httpappoi

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, h *Handler) {
	r.POST("/appointments", h.CreateAppointment)
	r.GET("/appointments/:id", h.GetByID)
	r.GET("/appointments", h.GetAll)
	r.PATCH("/appointments/:id/status", h.UpdateStatus)
}
