package repo

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"nettest/internal/models"
	"strconv"
)

type Repo struct {
	db *gorm.DB
}

func NewRepository(connectionString string) (*Repo, error) {
	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	//err = db.AutoMigrate(&models.Modeling{})
	//if err != nil {
	//	panic("cant migrate db")
	//}

	return &Repo{
		db: db,
	}, nil
}

func (r *Repo) GetModelingByID(modelingId int) (models.Modeling, error) {
	modeling := models.Modeling{}

	err := r.db.First(&modeling, "modeling_id = ?", strconv.Itoa(modelingId)).Error
	if err != nil {
		return modeling, err
	}

	return modeling, nil
}

func (r *Repo) DeleteModelingByID(modelingId int) error {
	err := r.db.Exec("UPDATE modelings SET is_deleted=true WHERE modeling_id = ?", modelingId).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *Repo) GetModelings(from, to int) ([]models.Modeling, error) {
	modelings := make([]models.Modeling, 0)

	r.db.Where("is_deleted = ? AND price >= ? AND ? >= price", false, from, to).Find(&modelings)

	return modelings, nil
}
