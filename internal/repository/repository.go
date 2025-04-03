package repository

import (
	"fmt"
	"log/slog"

	"github.com/grigory222/scraptor/internal/config"
	"github.com/grigory222/scraptor/internal/model"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Postgres struct {
	DB  *sqlx.DB
	log *slog.Logger
}

func NewPostgres(cfg config.DBConfig, log *slog.Logger) *Postgres {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName)
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Error(err.Error())
	}
	return &Postgres{DB: db, log: log}
}

// ================= Chats =================

func (p *Postgres) AddChat(id int) error {
	// на данный момент только лс с ботом
	chatType := "personal"

	query := `INSERT INTO chats (id, type) VALUES ($1, $2)`

	_, err := p.DB.Exec(query, id, chatType)
	if err != nil {
		return err
	}

	return nil
}

// ================= Links =================

// Дублирование элементов сделано намерено,
// поскольку за каждой ссылкой закреплен tokenID,
// который у каждого свой соответственно

func (p *Postgres) AddLink(link string, tag string, tokenID int, chatID int) (*model.Link, error) {
	// Начинаем транзакцию
	tx, err := p.DB.Beginx()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	// Вставляем запись в таблицу links
	query := `INSERT INTO links (link, tag, token_id) VALUES ($1, $2, $3) RETURNING id`
	var newID int
	if tokenID == 0 {
		err = tx.Get(&newID, query, link, tag, nil)
	} else {
		err = tx.Get(&newID, query, link, tag, tokenID)
	}
	if err != nil {
		return nil, err
	}

	// Вставляем запись в таблицу chats_links
	insertChatLinkQuery := `INSERT INTO chats_links (chat_id, link_id) VALUES ($1, $2)`
	_, err = tx.Exec(insertChatLinkQuery, chatID, newID)
	if err != nil {
		return nil, err
	}

	// Если все успешно, коммитим транзакцию
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return model.NewLink(newID, link, tag, tokenID), nil
}
