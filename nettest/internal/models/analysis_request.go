package models

import "time"

type AnalysisRequest struct {
	RequestId     int       `gorm:"primaryKey" json:"requestId"`
	UserId        int       `json:"userId"`
	User          string    `json:"user"`
	Status        string    `json:"status"`
	CreationDate  time.Time `json:"creationDate"`
	FormationDate time.Time `json:"formationDate"`
	CompleteDate  time.Time `json:"completeDate"`
	AdminId       int       `gorm:"default:null" json:"adminId"`
	Admin         string    `json:"admin"`
}

type RequestCreateMessage struct {
	UserId     int `json:"userId"`
	ModelingId int `json:"modelingId"`
	RequestId  int `json:"requestId"`
}

type ModelingInRequestMessage struct {
	ModelingId     int     `json:"modelingId"`
	Name           string  `json:"name"`
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
