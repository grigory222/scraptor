package service

import (
	"errors"
	"fmt"
	"testing"

	"github.com/grigory222/scraptor/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	one  = 1
	zero = 0
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) AddChat(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRepository) DeleteTgChat(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRepository) AddLink(link, tag string, tokenID, chatID int) (*model.Link, error) {
	args := m.Called(link, tag, tokenID, chatID)
	linkk := args.Get(0)
	if linkk != nil {
		return linkk.(*model.Link), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRepository) GetLinks(chatID int) ([]model.Link, error) {
	args := m.Called(chatID)
	links := args.Get(0)
	if links != nil {
		return links.([]model.Link), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRepository) DeleteLink(chatID int, link string) (*model.Link, error) {
	args := m.Called(chatID, link)
	linkk := args.Get(0)
	if linkk != nil {
		return linkk.(*model.Link), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestAddTgChat(t *testing.T) {
	tests := []struct {
		name        string
		chatID      int
		mockSetup   func(*MockRepository)
		expectedErr error
	}{
		{
			name:   "success",
			chatID: 123,
			mockSetup: func(m *MockRepository) {
				m.On("AddChat", 123).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:   "duplicate chat",
			chatID: 456,
			mockSetup: func(m *MockRepository) {
				m.On("AddChat", 456).Return(errors.New("pq: duplicate key value violates unique constraint"))
			},
			expectedErr: errors.New("pq: duplicate key value violates unique constraint"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			tt.mockSetup(repo)

			s := NewService(repo, nil)
			err := s.AddTgChat(tt.chatID)

			assert.Equal(t, tt.expectedErr, err)
			repo.AssertExpectations(t)
		})
	}
}

func TestDeleteTgChat(t *testing.T) {
	tests := []struct {
		name        string
		chatID      int
		mockSetup   func(*MockRepository)
		expectedErr error
	}{
		{
			name:   "success",
			chatID: 123,
			mockSetup: func(m *MockRepository) {
				m.On("DeleteTgChat", 123).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:   "chat not found",
			chatID: 456,
			mockSetup: func(m *MockRepository) {
				m.On("DeleteTgChat", 456).Return(fmt.Errorf("nothing deleted"))
			},
			expectedErr: fmt.Errorf("nothing deleted"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			tt.mockSetup(repo)

			s := NewService(repo, nil)
			err := s.DeleteTgChat(tt.chatID)

			assert.Equal(t, tt.expectedErr, err)
			repo.AssertExpectations(t)
		})
	}
}

func TestAddLink(t *testing.T) {
	tests := []struct {
		name        string
		chatID      int
		link        model.LinkRequestDTO
		mockSetup   func(*MockRepository)
		expected    *model.Link
		expectedErr error
	}{
		{
			name:   "success",
			chatID: 123,
			link:   model.LinkRequestDTO{Link: "https://example.com", Tag: "test", TokenID: 1},
			mockSetup: func(m *MockRepository) {
				m.On("AddLink", "https://example.com", "test", 1, 123).
					Return(&model.Link{ID: 1, Link: "https://example.com", Tag: "test", TokenID: &one}, nil)
			},
			expected:    &model.Link{ID: 1, Link: "https://example.com", Tag: "test", TokenID: &one},
			expectedErr: nil,
		},
		{
			name:   "repository error",
			chatID: 123,
			link:   model.LinkRequestDTO{Link: "https://error.com", Tag: "test", TokenID: 1},
			mockSetup: func(m *MockRepository) {
				m.On("AddLink", "https://error.com", "test", 1, 123).
					Return(nil, errors.New("db error"))
			},
			expected:    nil,
			expectedErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			tt.mockSetup(repo)

			s := NewService(repo, nil)
			result, err := s.AddLink(tt.chatID, tt.link)

			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}
			assert.Equal(t, tt.expectedErr, err)
			repo.AssertExpectations(t)
		})
	}
}

func TestGetLinks(t *testing.T) {
	tests := []struct {
		name        string
		chatID      int
		mockSetup   func(*MockRepository)
		expected    []model.Link
		expectedErr error
	}{
		{
			name:   "success",
			chatID: 123,
			mockSetup: func(m *MockRepository) {
				m.On("GetLinks", 123).
					Return([]model.Link{
						{ID: 1, Link: "https://example.com", Tag: "test1", TokenID: &one},
						{ID: 2, Link: "https://example.org", Tag: "test2", TokenID: &zero},
					}, nil)
			},
			expected: []model.Link{
				{ID: 1, Link: "https://example.com", Tag: "test1", TokenID: &one},
				{ID: 2, Link: "https://example.org", Tag: "test2", TokenID: &zero},
			},
			expectedErr: nil,
		},
		{
			name:   "empty result",
			chatID: 456,
			mockSetup: func(m *MockRepository) {
				m.On("GetLinks", 456).Return([]model.Link{}, nil)
			},
			expected:    []model.Link{},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			tt.mockSetup(repo)

			s := NewService(repo, nil)
			result, err := s.GetLinks(tt.chatID)

			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expectedErr, err)
			repo.AssertExpectations(t)
		})
	}
}

func TestDeleteLink(t *testing.T) {
	tests := []struct {
		name        string
		chatID      int
		link        model.LinkDeleteRequestDTO
		mockSetup   func(*MockRepository)
		expected    *model.Link
		expectedErr error
	}{
		{
			name:   "success",
			chatID: 123,
			link:   model.LinkDeleteRequestDTO{Link: "https://example.com"},
			mockSetup: func(m *MockRepository) {
				m.On("DeleteLink", 123, "https://example.com").
					Return(&model.Link{ID: 1, Link: "https://example.com", Tag: "test", TokenID: &one}, nil)
			},
			expected:    &model.Link{ID: 1, Link: "https://example.com", Tag: "test", TokenID: &one},
			expectedErr: nil,
		},
		{
			name:   "not found",
			chatID: 123,
			link:   model.LinkDeleteRequestDTO{Link: "https://notfound.com"},
			mockSetup: func(m *MockRepository) {
				m.On("DeleteLink", 123, "https://notfound.com").
					Return(nil, errors.New("link not found"))
			},
			expected:    nil,
			expectedErr: errors.New("link not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			tt.mockSetup(repo)

			s := NewService(repo, nil)
			result, err := s.DeleteLink(tt.chatID, tt.link)

			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expectedErr, err)
			repo.AssertExpectations(t)
		})
	}
}
