package app

import (
	"nettest/internal/models"
	"time"
)

type Repo interface {
	GetModelings(from, to int) ([]models.Modeling, error)
	GetModelingByID(modelingId int) (models.Modeling, error)
	DeleteModelingByID(modelingId int) error
	GetDraftRequest(userId int) (int, error)
	AddModeling(newModeling models.Modeling) error
	GetModelingImage(modelingId int) string
	UpdateModeling(modeling models.Modeling) error
	AddModelingToRequest(modeling models.RequestCreateMessage) error
	GetAnalysisRequests(status string, startDate, endDate time.Time) ([]models.AnalysisRequest, error)
	GetAnalysisRequestById(id int) (models.AnalysisRequest, []models.ModelingInRequestMessage, error)
	UpdateAnalysisRequestStatus(requestId int, status string) error
	DeleteModelingFromRequest(userId, modelingId int) (models.AnalysisRequest, []models.ModelingInRequestMessage, error)
	UpdateModelingRequest(userId int, updateModelingRequest models.AnalysisRequestsModeling) error
}
