package handlers_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grigory222/scraptor/internal/http-server/handlers"
	"github.com/grigory222/scraptor/internal/model"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockService struct {
	mock.Mock
}

func (m *mockService) AddTgChat(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

// добавим пустые реализации других методов интерфейса
func (m *mockService) GetHello() string                                       { return "" }
func (m *mockService) DeleteTgChat(id int) error                              { return nil }
func (m *mockService) AddLink(int, model.LinkRequestDTO) (*model.Link, error) { return nil, nil }
func (m *mockService) DeleteLink(int, model.LinkDeleteRequestDTO) (*model.Link, error) {
	return nil, nil
}
func (m *mockService) GetLinks(int) ([]model.Link, error) {
	return nil, nil
}

func TestAddTgChat(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name           string
		param          string
		mockSetup      func(m *mockService)
		expectedStatus int
		expectError    bool
	}{
		{
			name:  "valid id",
			param: "123",
			mockSetup: func(m *mockService) {
				m.On("AddTgChat", 123).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name:           "invalid id (not a number)",
			param:          "abc",
			mockSetup:      func(m *mockService) {}, // не вызывается
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:  "service returns error",
			param: "456",
			mockSetup: func(m *mockService) {
				m.On("AddTgChat", 456).Return(errors.New("some error"))
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/tg-chat/"+tt.param, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.param)

			svc := new(mockService)
			if tt.mockSetup != nil {
				tt.mockSetup(svc)
			}

			h := handlers.NewHandler(svc)

			err := h.AddTgChat(c)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedStatus != 0 {
					httpErr, ok := err.(*echo.HTTPError)
					assert.True(t, ok)
					assert.Equal(t, tt.expectedStatus, httpErr.Code)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)
			}

			svc.AssertExpectations(t)
		})
	}
}

func TestDeleteTgChat(t *testing.T) {
	type testCase struct {
		name       string
		param      string
		mockSetup  func(m *mockService)
		wantStatus int
		wantBody   string
	}

	tests := []testCase{
		{
			name:  "valid id",
			param: "123",
			mockSetup: func(m *mockService) {
				m.On("DeleteTgChat", 123).Return(nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   "\"\"\n",
		},
		{
			name:  "invalid id (not a number)",
			param: "abc",
			mockSetup: func(m *mockService) {
				// не нужен мок, ошибка произойдёт на уровне парсинга
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "strconv.Atoi: parsing \"abc\": invalid syntax\n",
		},
		{
			name:  "id not found in service",
			param: "456",
			mockSetup: func(m *mockService) {
				m.On("DeleteTgChat", 456).Return(errors.New("not found"))
			},
			wantStatus: http.StatusNotFound,
			wantBody:   "{\"message\":\"Couldn't delete tg-chat with such id: 456\\nError: not found\"}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, "/tg-chat/"+tt.param, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.param)

			mockSvc := new(mockService)
			if tt.mockSetup != nil {
				tt.mockSetup(mockSvc)
			}
			h := handlers.NewHandler(mockSvc)

			err := h.DeleteTgChat(c)
			if tt.param == "456" {
				fmt.Print("hereeee")
				fmt.Print(err)
			}
			if err != nil {
				httpErr, ok := err.(*echo.HTTPError)
				if ok {
					assert.Equal(t, tt.wantStatus, httpErr.Code)
					assert.Equal(t, tt.wantBody, httpErr.Message)
				} else {
					// ошибка парсинга id, возвращается напрямую
					assert.Equal(t, tt.wantBody, err.Error()+"\n")
				}
			} else {
				assert.Equal(t, tt.wantStatus, rec.Code)
				assert.Equal(t, tt.wantBody, rec.Body.String())
			}

			mockSvc.AssertExpectations(t)
		})
	}
}
