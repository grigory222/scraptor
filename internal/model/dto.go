package model

type LinkRequestDTO struct {
	ID      int    `json:"id"`
	Link    string `json:"link"`
	Tag     string `json:"tag"`
	TokenID int    `json:"token_id"`
}

type LinkResponseDTO struct {
	ID      int    `json:"id"`
	Link    string `json:"link"`
	Tag     string `json:"tag"`
	TokenID int    `json:"token_id"`
}
