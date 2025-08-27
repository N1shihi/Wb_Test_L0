package handler

import (
	"Wb_Test_L0/internal/models"
	"Wb_Test_L0/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type Handler struct {
	srv *service.Service
}

func New(srv *service.Service) *Handler {
	return &Handler{srv: srv}
}

func (h *Handler) SaveOrder(order models.Order) error {
	return h.srv.SaveOrder(order)
}

func (h *Handler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order id required"})
		return
	}
	start := time.Now()
	order, err := h.srv.GetOrder(id)
	duration := time.Since(start)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":       err.Error(),
			"duration_ms": duration.Milliseconds(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"order":       order,
		"duration_us": duration.Microseconds(),
	})
}

func (h *Handler) GetHtml(c *gin.Context) {
	c.File("./web/index.html")
}
