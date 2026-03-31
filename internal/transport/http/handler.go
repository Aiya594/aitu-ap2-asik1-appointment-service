package httpappoi

import (
	"net/http"

	"github.com/Aiya594/appointment-services/internal/model"
	usecase "github.com/Aiya594/appointment-services/internal/use-case"
	"github.com/gin-gonic/gin"
)

// Appointment Service Endpoints
//  POST /appointments - create a new appointment
//  GET /appointments/{id} - retrieve an appointment by ID
//  GET /appointments - list all appointments
//  PATCH /appointments/{id}/status - update appointment status

type Handler struct {
	us usecase.AppointmentUseCase
}

type request struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	DoctorID    string `json:"doctor_id"`
}

type statusRequest struct {
	Status model.Status `json:"status"`
}

func NewAppointmentHandler(us usecase.AppointmentUseCase) *Handler {
	return &Handler{us: us}
}

func (h *Handler) CreateAppointment(c *gin.Context) {
	var req request

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(parseError(err), gin.H{
			"error": "invalid request body",
		})
		return
	}

	err := h.us.CreateAppointment(req.Title, req.Description, req.DoctorID)
	if err != nil {
		c.JSON(parseError(err), gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "appointment created",
	})
}

func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")

	ap, err := h.us.GetByID(id)
	if err != nil {
		c.JSON(parseError(err), gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ap)
}

func (h *Handler) GetAll(c *gin.Context) {
	list, err := h.us.GetAll()
	if err != nil {
		c.JSON(parseError(err), gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, list)
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")

	var req statusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(parseError(err), gin.H{
			"error": "invalid request body",
		})
		return
	}

	err := h.us.UpdateStatus(id, req.Status)
	if err != nil {
		c.JSON(parseError(err), gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "status updated",
	})
}
