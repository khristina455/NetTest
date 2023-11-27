package models

type Modeling struct {
	ModelingId  int `gorm:"primaryKey"`
	Name        string
	Description string
	Image       string
	IsDeleted   bool
	Price       float32
}
