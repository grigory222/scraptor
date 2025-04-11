package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/grigory222/scraptor/internal/http-server/middlewares"
	"github.com/grigory222/scraptor/internal/model"
	"github.com/grigory222/scraptor/internal/service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Handler struct {
	service service.IService
}

// NewHandler создаёт новый хендлер
func NewHandler(svc service.IService) *Handler {
	return &Handler{service: svc}
}

// RegisterRoutes регистрирует маршруты
func RegisterRoutes(e *echo.Echo, svc service.IService) {
	h := NewHandler(svc)
	e.POST("/tg-chat/:id", h.AddTgChat)
	e.DELETE("/tg-chat/:id", h.DeleteTgChat)
	e.POST("/links", h.AddLink)
	e.GET("/links", h.GetLinks)
	e.DELETE("/links", h.DeleteLink)
}

func RegisterMiddlewares(e *echo.Echo) {
	e.Use(middleware.RequestID())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))
	e.Use(middlewares.ErrorHandlerMiddleware)
}

func (h *Handler) AddTgChat(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	err = h.service.AddTgChat(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Couldn't add tg-chat with such id: %d\nError: %s", id, err.Error()))
	}
	return c.JSON(http.StatusCreated, "")
}

func (h *Handler) DeleteTgChat(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}
	err = h.service.DeleteTgChat(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("Couldn't delete tg-chat with such id: %d\nError: %s", id, err.Error()))
	}
	return c.JSON(http.StatusOK, "")
}

// ============= Links =============

func ValidateTgChatHeader(c echo.Context) (int, *echo.HTTPError) {
	// 	Tg-Chat-Id from header
	stringChatID := c.Request().Header.Get("Tg-Chat-Id")
	if stringChatID == "" {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "no header `Tg-Chat-Id` provided")
	}
	chatID, err := strconv.Atoi(stringChatID)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "incorrect value of header `Tg-Chat-Id` provided")
	}
	return chatID, nil
}

func (h *Handler) AddLink(c echo.Context) error {
	var linkReq model.LinkRequestDTO
	if err := c.Bind(&linkReq); err != nil {
		return err
	}

	if linkReq.Link == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "link field is required")
	}

	chatID, httpErr := ValidateTgChatHeader(c)
	if httpErr != nil {
		return httpErr
	}

	linkDAO, err := h.service.AddLink(chatID, linkReq)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if linkDAO == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "such link already exists")
	}
	linkResp := linkDAO.ToResponseDTO()
	return c.JSON(http.StatusCreated, linkResp)
}

func (h *Handler) DeleteLink(c echo.Context) error {
	var linkReq model.LinkDeleteRequestDTO
	if err := c.Bind(&linkReq); err != nil {
		return err
	}

	chatID, httpErr := ValidateTgChatHeader(c)
	if httpErr != nil {
		return httpErr
	}

	linkDAO, err := h.service.DeleteLink(chatID, linkReq)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	linkResp := linkDAO.ToResponseDTO()
	return c.JSON(http.StatusOK, linkResp)
}

func (h *Handler) GetLinks(c echo.Context) error {
	chatID, httpErr := ValidateTgChatHeader(c)
	if httpErr != nil {
		return httpErr
	}

	linksDAO, err := h.service.GetLinks(chatID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// convert to response DTO
	linksResponse := make([]model.LinkResponseDTO, len(linksDAO))
	for i, link := range linksDAO {
		linksResponse[i] = *link.ToResponseDTO()
	}

	return c.JSON(http.StatusOK, linksResponse)
}
