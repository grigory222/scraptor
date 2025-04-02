package handlers

import (
	"net/http"

	"github.com/grigory222/scraptor/src/service"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	service *service.Service
}

// NewHandler создаёт новый хендлер
func NewHandler(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

// RegisterRoutes регистрирует маршруты
func RegisterRoutes(e *echo.Echo, svc *service.Service) {
	h := NewHandler(svc)
	e.GET("/hello", h.Hello)
}

func (h *Handler) Hello(c echo.Context) error {
	return c.HTML(http.StatusOK, h.service.GetHello())
}

// func (h *Handler) GetItems(c echo.Context) error {
// 	items, err := h.service.GetItems()
// 	if err != nil {
// 		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
// 	}
// 	return c.JSON(http.StatusOK, items)
// }
