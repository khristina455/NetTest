package models

type Modeling struct {
	ModelingId  int     `gorm:"primaryKey" json:"modelingId"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Image       string  `json:"image"`
	IsDeleted   bool    `json:"isDeleted"`
	Price       float32 `json:"price"`
}
