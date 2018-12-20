package models

//easyjson:json
type Store struct {
	Coins          *int   `json:"coins,omitempty"`
	PurchasedSkins []uint `json:"skins,omitempty"`
	Skin           *uint  `json:"current_skin,omitempty"`
}

//easyjson:json
type Skin struct {
	ID   uint   `json:"id" db:"skin_id"`
	Name string `json:"name" db:"skin_name"`
	Cost int    `json:"cost"`
}

//easyjson:json
type AllSkins struct {
	Skins []Skin `json:"skins"`
}

//easyjson:json
type RequestSkin struct {
	ID uint `json:"skin"`
}
