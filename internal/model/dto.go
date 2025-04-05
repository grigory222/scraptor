package model

type LinkRequestDTO struct {
	Link    string `json:"link"`
	Tag     string `json:"tag"`
	TokenID int    `json:"token_id"`
}

type LinkDeleteRequestDTO struct {
	Link string `json:"link"`
}

type LinkResponseDTO struct {
	ID      int    `json:"id"`
	Link    string `json:"link"`
	Tag     string `json:"tag"`
	TokenID int    `json:"token_id"`
}
