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
}

type AnalysisRequestsModeling struct {
	RequestId      int `gorm:"primaryKey"`
	ModelingId     int `gorm:"primaryKey"`
	NodeQuantity   int
	QueueSize      int
	ClientQuantity int
}
