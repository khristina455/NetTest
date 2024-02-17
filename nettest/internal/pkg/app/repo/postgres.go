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

func (r *Repo) GetRequestById(requestId int) (models.AnalysisRequest, []models.ModelingInRequestMessage, error) {
	var analysisRequest models.AnalysisRequest
	var modelings []models.Modeling
	var requestModeling []models.AnalysisRequestsModeling
	var modelingsWithFields []models.ModelingInRequestMessage

	result := r.db.First(&analysisRequest, "request_id =?", requestId)
	if result.Error != nil {
		return models.AnalysisRequest{}, nil, result.Error
	}

	res := r.db.
		Table("analysis_requests_modelings").
		Select("modelings.*").
		Joins("JOIN modelings ON analysis_requests_modelings.modeling_id = modelings.modeling_id").
		Where("analysis_requests_modelings.request_id = ?", requestId).
		Find(&modelings)
	if res.Error != nil {
		return models.AnalysisRequest{}, nil, res.Error
	}

	res = r.db.Where("request_id = ?", requestId).Find(&requestModeling)
	if res.Error != nil {
		return models.AnalysisRequest{}, nil, res.Error
	}

	for ind := range modelings {
		var currentRequest models.ModelingInRequestMessage
		currentRequest.ModelingId = modelings[ind].ModelingId
		currentRequest.Name = modelings[ind].Name
		currentRequest.Image = modelings[ind].Image
		currentRequest.Description = modelings[ind].Description
		currentRequest.IsDeleted = modelings[ind].IsDeleted
		currentRequest.Price = modelings[ind].Price

		currentRequest.ClientQuantity = requestModeling[ind].ClientQuantity
		currentRequest.QueueSize = requestModeling[ind].QueueSize
		currentRequest.NodeQuantity = requestModeling[ind].NodeQuantity
		currentRequest.Result = requestModeling[ind].Result

		modelingsWithFields = append(modelingsWithFields, currentRequest)
	}

	return analysisRequest, modelingsWithFields, nil
}
