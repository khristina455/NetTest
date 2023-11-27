package models

type User struct {
	UserId   int `gorm:"primaryKey"`
	Login    string
	IsAdmin  bool
	Name     string
	Password string
}
