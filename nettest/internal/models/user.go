package models

type User struct {
	UserId   int `gorm:"primaryKey"`
	Login    string
	IsAdmin  bool
	Name     string
	Password string
}

func GetClientId() int {
	return 1
}

func GetAdminId() int {
	return 2
}
