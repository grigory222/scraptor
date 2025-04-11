package repository

import (
	"github.com/grigory222/scraptor/internal/model"
)

type Repository interface {
	AddChat(id int) error
	DeleteTgChat(id int) error
	AddLink(link, tag string, tokenID, chatID int) (*model.Link, error)
	GetLinks(chatID int) ([]model.Link, error)
	DeleteLink(chatID int, link string) (*model.Link, error)
}
