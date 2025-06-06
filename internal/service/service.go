package service

import (
	"io"
	"log/slog"

	"github.com/grigory222/scraptor/internal/model"
	"github.com/grigory222/scraptor/internal/repository"
)

type Service struct {
	db  repository.Repository
	log *slog.Logger
}

func NewService(db repository.Repository, log *slog.Logger) *Service {
	if log == nil {
		log = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &Service{db: db, log: log}
}

func (s *Service) AddTgChat(id int) error {
	err := s.db.AddChat(id)
	if err != nil && s.log != nil {
		s.log.Error(err.Error())
	}

	return err
}

func (s *Service) DeleteTgChat(id int) error {
	err := s.db.DeleteTgChat(id)
	if err != nil {
		if s.log != nil {
			s.log.Error(err.Error())
		}

		return err
	}
	return nil
}

func (s *Service) AddLink(chatID int, link model.LinkRequestDTO) (*model.Link, error) {
	linkDAO, err := s.db.AddLink(link.Link, link.Tag, link.TokenID, chatID)
	if err != nil {
		return nil, err
	}
	return linkDAO, nil
}

func (s *Service) GetLinks(chatID int) ([]model.Link, error) {
	linksDAO, err := s.db.GetLinks(chatID)
	if err != nil {
		return nil, err
	}
	return linksDAO, nil
}

func (s *Service) DeleteLink(chatID int, link model.LinkDeleteRequestDTO) (*model.Link, error) {
	linkDeleted, err := s.db.DeleteLink(chatID, link.Link)
	if err != nil {
		if s.log != nil {
			s.log.Error(err.Error())
		}
		return nil, err
	}
	return linkDeleted, nil
}
