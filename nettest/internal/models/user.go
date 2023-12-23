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

type UserLogin struct {
	Login    string `json:"login" binding:"required,max=64"`
	Password string `json:"password" binding:"required,min=8,max=64"`
}

type UserSignUp struct {
	Login    string `json:"login" binding:"required,max=64"`
	Name     string `json:"name"`
	Password string `json:"password" binding:"required,min=8,max=64"`
}
