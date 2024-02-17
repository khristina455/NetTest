package app

import (
	"nettest/internal/models"
	"time"
)

type Repo interface {
	GetModelings(query string, from, to int) ([]models.Modeling, error)
	GetModelingByID(modelingId int) (models.Modeling, error)
	DeleteModelingByID(modelingId int) error
	GetDraftRequest(userId int) (int, error)
	AddModeling(newModeling models.Modeling) error
	GetModelingImage(modelingId int) string
	UpdateModeling(modeling models.Modeling) error
	AddModelingToRequest(modeling models.RequestCreateMessage) error
	GetAnalysisRequests(status string, startDate, endDate time.Time, userId int, isAdmin bool) ([]models.AnalysisRequest, error)
	GetAnalysisRequestById(id int, userId int, isAdmin bool) (models.AnalysisRequest, []models.ModelingInRequestMessage, error)
	DeleteAnalysisRequest(userId int) error
	UpdateAnalysisRequestStatusAdmin(adminId int, reqId int, status string) error
	UpdateAnalysisRequestStatusClient(userId int, status string) (int, error)
	DeleteModelingFromRequest(userId, modelingId int) (models.AnalysisRequest, []models.ModelingInRequestMessage, error)
	UpdateModelingRequest(userId int, updateModelingRequest models.AnalysisRequestsModeling) error
	SignUp(newUser models.User) error
	GetByCredentials(user models.User) (models.User, error)
	GetUserInfo(user models.User) (models.User, error)
	WriteResult(requestId int, modelingId int, result int) error
	GetStatisticsForRequests() []models.StatisticMessage
}
