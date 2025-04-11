package repository

import (
	"database/sql"
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

// func (p *Postgres) GetTgChat(id int) (*model.Chat, error) {
// 	query := `SELECT id, type FROM chats WHERE id = $1`
// 	chat := model.Chat{}
// 	err := p.DB.Get(&chat, query, id)
// 	if err != nil {
// 		p.log.Error("Can't select chat", "id", id, "err", err)
// 		return nil, err
// 	}
// 	return &chat, nil
// }

func (p *Postgres) DeleteTgChat(id int) error {
	query := `DELETE FROM chats WHERE id = $1`
	res, err := p.DB.Exec(query, id)
	if err != nil {
		p.log.Error(err.Error())
		return err
	}
	rows, _ := res.RowsAffected()
	if rows < 1 {
		return fmt.Errorf("nothing deleted")
	}
	return nil
}

// ================= Links =================

// Дублирование элементов сделано намерено,
// поскольку за каждой ссылкой закреплен tokenID,
// который у каждого свой соответственно

func (p *Postgres) AddLink(link string, tag string, tokenID int, chatID int) (*model.Link, error) {
	linkFound, err := p.GetLink(chatID, link)
	if err != nil || linkFound != nil {
		return nil, err
	}
	p.log.Debug("here0")

	// Начинаем транзакцию
	tx, err := p.DB.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Вставляем запись в таблицу links
	query := `INSERT INTO links (link, tag, token_id) VALUES ($1, $2, $3) RETURNING id`
	var newID int
	p.log.Debug("", "tokenID", tokenID)
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

func (p *Postgres) GetLinks(chatID int) ([]model.Link, error) {
	query := `SELECT links.id, links.link, links.tag, links.token_id FROM links 
			  JOIN chats_links on links.id = chats_links.link_id
			  WHERE chats_links.chat_id = $1`
	var links []model.Link
	err := p.DB.Select(&links, query, chatID)
	if err != nil {
		return nil, err
	}

	return links, nil
}

func (p *Postgres) GetLink(chatID int, link string) (*model.Link, error) {
	query := `SELECT links.id, links.link, links.tag, links.token_id FROM links
			  JOIN chats_links cl on cl.link_id = links.id
			  WHERE cl.chat_id = $1 and links.link = $2`
	var linkRes model.Link
	err := p.DB.Get(&linkRes, query, chatID, link)
	p.log.Debug("herrr")
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &linkRes, nil
}

func (p *Postgres) DeleteLink(chatID int, link string) (*model.Link, error) {
	linkFound, err := p.GetLink(chatID, link)

	if err != nil {
		return nil, err
	}
	if linkFound == nil {
		return nil, fmt.Errorf("not found such link")
	}

	// начать транзакцию
	tx, err := p.DB.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// удалить ссылку
	query := `DELETE FROM links
			  WHERE links.id = $1`
	_, err = tx.Exec(query, linkFound.ID)
	if err != nil {
		return nil, err
	}

	// удалить её токен
	query = `DELETE FROM tokens
			  WHERE tokens.id = $1`
	if linkFound.TokenID != nil {
		_, err := tx.Exec(query, *linkFound.TokenID)
		if err != nil {
			return nil, err
		}
	}

	// завершить транзакцию
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return linkFound, nil
}
