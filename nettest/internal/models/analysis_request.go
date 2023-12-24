package models

import "time"

type AnalysisRequest struct {
	RequestId     int `gorm:"primaryKey"`
	UserId        int
	Status        string
	CreationDate  time.Time
	FormationDate time.Time
	CompleteDate  time.Time
	AdminId       int `gorm:"default:null"`
}

type RequestCreateMessage struct {
	UserId     int `json:"userId"`
	ModelingId int `json:"modelingId"`
	RequestId  int `json:"requestId"`
}

type ModelingInRequestMessage struct {
	ModelingId     int     `json:"modelingId"`
	Name           string  `json:"name""`
	Description    string  `json:"description"`
	Image          string  `json:"image"`
	IsDeleted      bool    `json:"isDeleted"`
	Price          float32 `json:"price"`
	NodeQuantity   int     `json:"nodeQuantity"`
	QueueSize      int     `json:"queueSize"`
	ClientQuantity int     `json:"clientQuantity"`
}

type AnalysisRequestsModeling struct {
	RequestId      int `gorm:"primaryKey"`
	ModelingId     int `gorm:"primaryKey"`
	NodeQuantity   int
	QueueSize      int
	ClientQuantity int
}
