package service

import "github.com/grigory222/scraptor/src/repository"

type Service struct {
	db *repository.Postgres
}

func NewService(db *repository.Postgres) *Service {
	return &Service{db: db}
}

// логика работы
// ...

func (s *Service) GetHello() string {
	return "Hello, world!"
}
