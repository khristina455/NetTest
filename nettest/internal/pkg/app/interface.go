package app

import "nettest/internal/models"

type Repo interface {
	GetModelings(from, to int) ([]models.Modeling, error)
	GetModelingByID(modelingId int) (models.Modeling, error)
	DeleteModelingByID(modelingId int) error
	GetRequestById(id int) (models.AnalysisRequest, []models.ModelingInRequestMessage, error)
}
