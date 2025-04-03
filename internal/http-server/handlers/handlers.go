package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/grigory222/scraptor/internal/http-server/middlewares"
	"github.com/grigory222/scraptor/internal/logger"
	"github.com/grigory222/scraptor/internal/model"
	"github.com/grigory222/scraptor/internal/service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
	e.POST("/tg-chat/:id", h.AddTgChat)
	e.POST("/links", h.AddLink)
}

func RegisterMiddlewares(e *echo.Echo) {
	e.Use(middleware.RequestID())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))
	e.Use(middlewares.ErrorHandlerMiddleware)
}

func (h *Handler) Hello(c echo.Context) error {
	return c.HTML(http.StatusOK, h.service.GetHello())
}

func (h *Handler) AddTgChat(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}
	err = h.service.AddTgChat(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Couldn't add tg-chat with such id: %d\nError: %s", id, err.Error()))
	}
	return c.JSON(http.StatusCreated, "")
}

func (h *Handler) AddLink(c echo.Context) error {
	var linkReq model.LinkRequestDTO
	if err := c.Bind(&linkReq); err != nil {
		return err
	}

	// 	Tg-Chat-Id from header
	stringChatID := c.Request().Header.Get("Tg-Chat-Id")
	if stringChatID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "no header `Tg-Chat-Id` provided")
	}
	chatID, err := strconv.Atoi(stringChatID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "incorrect value of header `Tg-Chat-Id` provided")
	}

	linkResp, err := h.service.AddLink(chatID, linkReq)
	if err != nil {
		logger.Logger.Debug("error in add link handler")
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, linkResp)
}
