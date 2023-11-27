package app

import "nettest/internal/models"

type Repo interface {
	GetModelings() ([]models.Modeling, error)
	GetModelingByID(modelingId int) (models.Modeling, error)
	DeleteModelingByID(modelingId int) error
}
