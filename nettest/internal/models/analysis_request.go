package models

import "time"

type AnalysisRequest struct {
	RequestId      int `gorm:"primaryKey"`
	UserId         int
	Status         string
	CreationDate   time.Time
	FormationDate  time.Time
	CompleteDate   time.Time
	NodeQuantity   int
	QueueSize      int
	ClientQuantity int
	Result         float32
	AdminId        int
}
