package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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

// добавим пустые реализации других методов интерфейса
func (m *mockService) AddTgChat(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *mockService) DeleteTgChat(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *mockService) AddLink(userID int, req model.LinkRequestDTO) (*model.Link, error) {
	args := m.Called(userID, req)
	return args.Get(0).(*model.Link), args.Error(1)
}

func (m *mockService) DeleteLink(userID int, req model.LinkDeleteRequestDTO) (*model.Link, error) {
	args := m.Called(userID, req)
	return args.Get(0).(*model.Link), args.Error(1)
}

func (m *mockService) GetLinks(userID int) ([]model.Link, error) {
	args := m.Called(userID)
	return args.Get(0).([]model.Link), args.Error(1)
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
			wantBody:   "Couldn't delete tg-chat with such id: 456\nError: not found",
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
			if err != nil {
				var httpErr *echo.HTTPError
				ok := errors.As(err, &httpErr)
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

func TestValidateTgChatHeader(t *testing.T) {
	tests := []struct {
		name        string
		headerValue string
		wantChatID  int
		wantError   *echo.HTTPError
	}{
		{
			name:        "valid header",
			headerValue: "123",
			wantChatID:  123,
			wantError:   nil,
		},
		{
			name:        "missing header",
			headerValue: "",
			wantChatID:  0,
			wantError:   echo.NewHTTPError(http.StatusBadRequest, "no header `Tg-Chat-Id` provided"),
		},
		{
			name:        "invalid header value",
			headerValue: "abc",
			wantChatID:  0,
			wantError:   echo.NewHTTPError(http.StatusBadRequest, "incorrect value of header `Tg-Chat-Id` provided"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.headerValue != "" {
				req.Header.Set("Tg-Chat-Id", tt.headerValue)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			chatID, err := handlers.ValidateTgChatHeader(c)

			assert.Equal(t, tt.wantChatID, chatID)
			assert.Equal(t, tt.wantError, err)
		})
	}
}

func TestAddLink(t *testing.T) {
	tests := []struct {
		name         string
		headerValue  string
		requestBody  string
		mockSetup    func(m *mockService)
		wantStatus   int
		wantResponse string
	}{
		{
			name:        "success with token",
			headerValue: "123",
			requestBody: `{"link": "https://example.com", "tag": "test", "token_id": 1}`,
			mockSetup: func(m *mockService) {
				m.On("AddLink", 123, model.LinkRequestDTO{
					Link:    "https://example.com",
					Tag:     "test",
					TokenID: 1,
				}).Return(model.NewLink(1, "https://example.com", "test", 1), nil)
			},
			wantStatus:   http.StatusCreated,
			wantResponse: `{"id":1,"link":"https://example.com","tag":"test","token_id":1}`,
		},
		{
			name:        "invalid request - missing link",
			headerValue: "123",
			requestBody: `{"tag": "test"}`,
			mockSetup:   func(m *mockService) {},
			wantStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			if tt.headerValue != "" {
				req.Header.Set("Tg-Chat-Id", tt.headerValue)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mockSvc := new(mockService)
			if tt.mockSetup != nil {
				tt.mockSetup(mockSvc)
			}
			h := handlers.NewHandler(mockSvc)

			err := h.AddLink(c)

			if tt.wantStatus >= 400 {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantStatus, rec.Code)
				assert.JSONEq(t, tt.wantResponse, rec.Body.String())
			}

			if tt.mockSetup != nil {
				mockSvc.AssertExpectations(t)
			}
		})
	}
}

func TestDeleteLink(t *testing.T) {
	tests := []struct {
		name         string
		headerValue  string
		requestBody  string
		mockSetup    func(m *mockService)
		wantStatus   int
		wantResponse string
	}{
		{
			name:        "success",
			headerValue: "123",
			requestBody: `{"link": "https://example.com"}`,
			mockSetup: func(m *mockService) {
				m.On("DeleteLink", 123, model.LinkDeleteRequestDTO{
					Link: "https://example.com",
				}).Return(model.NewLink(1, "https://example.com", "test", 1), nil)
			},
			wantStatus:   http.StatusOK,
			wantResponse: `{"id":1,"link":"https://example.com","tag":"test","token_id":1}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			if tt.headerValue != "" {
				req.Header.Set("Tg-Chat-Id", tt.headerValue)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mockSvc := new(mockService)
			if tt.mockSetup != nil {
				tt.mockSetup(mockSvc)
			}
			h := handlers.NewHandler(mockSvc)

			err := h.DeleteLink(c)

			if tt.wantStatus >= 400 {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantStatus, rec.Code)
				assert.JSONEq(t, tt.wantResponse, rec.Body.String())
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

func TestGetLinks(t *testing.T) {
	tests := []struct {
		name         string
		headerValue  string
		mockSetup    func(m *mockService)
		wantStatus   int
		wantResponse string
	}{
		{
			name:        "success with links",
			headerValue: "123",
			mockSetup: func(m *mockService) {
				m.On("GetLinks", 123).Return([]model.Link{
					*model.NewLink(1, "https://example.com", "test1", 1),
					*model.NewLink(2, "https://example.org", "test2", 0),
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantResponse: `[
                {"id":1,"link":"https://example.com","tag":"test1","token_id":1},
                {"id":2,"link":"https://example.org","tag":"test2","token_id":0}
            ]`,
		},
		{
			name:        "empty list",
			headerValue: "123",
			mockSetup: func(m *mockService) {
				m.On("GetLinks", 123).Return([]model.Link{}, nil)
			},
			wantStatus:   http.StatusOK,
			wantResponse: `[]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.headerValue != "" {
				req.Header.Set("Tg-Chat-Id", tt.headerValue)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mockSvc := new(mockService)
			if tt.mockSetup != nil {
				tt.mockSetup(mockSvc)
			}
			h := handlers.NewHandler(mockSvc)

			err := h.GetLinks(c)

			if tt.wantStatus >= 400 {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantStatus, rec.Code)
				assert.JSONEq(t, tt.wantResponse, rec.Body.String())
			}

			mockSvc.AssertExpectations(t)
		})
	}
}
