package service

import (
	"errors"
	"log/slog"

	"github.com/grigory222/scraptor/internal/model"
	"github.com/grigory222/scraptor/internal/repository"
)

type Service struct {
	db  *repository.Postgres
	log *slog.Logger
}

var (
	ErrNotFound      = errors.New("tg chat not found")
	ErrAlreadyExists = errors.New("tg chat already exists")
)

func NewService(db *repository.Postgres, log *slog.Logger) *Service {
	return &Service{db: db, log: log}
}

func (s *Service) GetHello() string {
	return "Hello, world!"
}

func (s *Service) AddTgChat(id int) error {
	err := s.db.AddChat(id)
	if err != nil {
		s.log.Error(err.Error())
	}

	return err
}

func (s *Service) AddLink(chatID int, link model.LinkRequestDTO) (*model.LinkResponseDTO, error) {
	linkDAO, err := s.db.AddLink(link.Link, link.Tag, link.TokenID, chatID)
	if err != nil {
		return nil, err
	}
	return linkDAO.ToResponseDTO(), nil
}
