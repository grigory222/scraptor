package service

import "github.com/grigory222/scraptor/internal/model"

type IService interface {
	AddTgChat(id int) error
	DeleteTgChat(id int) error
	AddLink(chatID int, req model.LinkRequestDTO) (*model.Link, error)
	DeleteLink(chatID int, req model.LinkDeleteRequestDTO) (*model.Link, error)
	GetLinks(chatID int) ([]model.Link, error)
}
