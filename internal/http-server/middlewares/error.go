package middlewares

import (
	"fmt"
	"net/http"
	"runtime"
	"strconv"

	"github.com/grigory222/scraptor/internal/logger"
	"github.com/labstack/echo/v4"
)

// APIError структура для универсальной обработки ошибок
type APIError struct {
	Description      string   `json:"description"`
	Code             string   `json:"code"`
	ExceptionName    string   `json:"exceptionName"`
	ExceptionMessage string   `json:"exceptionMessage"`
	Stacktrace       []string `json:"stacktrace"`
}

// NewAPIError создает новую ошибку с кодом и стектрейсом
func NewAPIError(description, code, exceptionName, exceptionMessage string) *APIError {
	stacktrace := getStacktrace()
	return &APIError{
		Description:      description,
		Code:             code,
		ExceptionName:    exceptionName,
		ExceptionMessage: exceptionMessage,
		Stacktrace:       stacktrace,
	}
}

// getStacktrace получает текущий стектрейс
func getStacktrace() []string {
	var stacktrace []string
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	for _, pc := range pc[:n] {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			file, line := fn.FileLine(pc)
			stacktrace = append(stacktrace, fmt.Sprintf("%s:%d", file, line))
		}
	}
	return stacktrace
}

// ErrorHandlerMiddleware для обработки ошибок и форматирования ответа
func ErrorHandlerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := next(c)
		if err != nil {
			var apiError *APIError

			code := http.StatusInternalServerError

			if he, ok := err.(*echo.HTTPError); ok {
				code = he.Code
				apiError = NewAPIError(
					he.Message.(string),
					strconv.Itoa(he.Code),
					"HTTPError",
					he.Message.(string),
				)
			} else {
				// Генерируем ошибку по умолчанию для других типов ошибок
				apiError = NewAPIError(
					"An unexpected error occurred",
					"500",
					"InternalServerError",
					err.Error(),
				)
			}

			logger.Logger.Error(apiError.Description)

			return c.JSON(code, apiError)
		}
		return nil
	}
}
