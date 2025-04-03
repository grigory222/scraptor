package model

type Link struct {
	ID      int    `db:"id"`
	Link    string `db:"link"`
	Tag     string `db:"tag"`
	TokenID int    `db:"token_id"`
}

type Chat struct {
	ID   int    `db:"id"`
	Type string `db:"type"`
}

func NewLink(id int, link, tag string, tokenID int) *Link {
	return &Link{id, link, tag, tokenID}
}

func (link *Link) ToResponseDTO() *LinkResponseDTO {
	return &LinkResponseDTO{link.ID, link.Link, link.Tag, link.TokenID}
}
