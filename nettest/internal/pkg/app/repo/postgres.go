package repo

import (
	"errors"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"nettest/internal/models"
	"strconv"
	"time"
)

type Repo struct {
	db *gorm.DB
}

func NewRepository(connectionString string) (*Repo, error) {
	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&models.Modeling{})
	if err != nil {
		log.Fatal("cant migrate db")
	}

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

func (r *Repo) GetDraftRequest(userId int) (int, error) {
	var analysisRequest models.AnalysisRequest
	err := r.db.First(&analysisRequest, "user_id = ? and status = 'DRAFT'", userId)
	if err.Error != nil && err.Error != gorm.ErrRecordNotFound {
		return 0, err.Error
	}

	return analysisRequest.RequestId, nil
}

func (r *Repo) AddModeling(newModeling models.Modeling) error {
	result := r.db.Create(&newModeling)
	return result.Error
}

func (r *Repo) GetModelingImage(modelingId int) string {
	modeling := models.Modeling{}

	r.db.First(&modeling, "modeling_id = ?", strconv.Itoa(modelingId))
	return modeling.Image
}

func (r *Repo) UpdateModeling(updateModeling models.Modeling) error {
	var modeling models.Modeling
	res := r.db.First(&modeling, "modeling_id =?", updateModeling.ModelingId)
	if res.Error != nil {
		return res.Error
	}

	if updateModeling.Name != "" {
		modeling.Name = updateModeling.Name
	}

	if updateModeling.Description != "" {
		modeling.Description = updateModeling.Description
	}

	if updateModeling.Image != "" {
		modeling.Image = updateModeling.Image
	}

	if updateModeling.Price != 0 {
		modeling.Price = updateModeling.Price
	}

	result := r.db.Save(modeling)
	return result.Error
}

func (r *Repo) AddModelingToRequest(modeling models.RequestCreateMessage) error {
	var request models.AnalysisRequest
	r.db.Where("user_id = ?", modeling.UserId).Where("status = ?", "DRAFT").First(&request)
	fmt.Println(request)

	if request.RequestId == 0 {
		newRequest := models.AnalysisRequest{
			UserId:       modeling.UserId,
			Status:       "DRAFT",
			CreationDate: time.Now(),
		}
		res := r.db.Create(&newRequest)
		if res.Error != nil {
			return res.Error
		}
		request = newRequest
	}

	requestsModelings := models.AnalysisRequestsModeling{
		RequestId:  request.RequestId,
		ModelingId: modeling.ModelingId,
	}

	res := r.db.Create(&requestsModelings)
	if res.Error != nil && res.Error.Error() == "ERROR: duplicate key value violates unique constraint \"monitoring_requests_threats_request_id_threat_id_key\" (SQLSTATE 23505)" {
		return errors.New("данная услуга уже добавлена в заявку")

	}

	return res.Error
}

func (r *Repo) GetAnalysisRequests(status string, startDate, endDate time.Time) ([]models.AnalysisRequest, error) {
	var analysisRequests []models.AnalysisRequest

	if status != "" {
		if startDate.IsZero() {
			if endDate.IsZero() {
				res := r.db.Where("status != ? AND status != ?", "DRAFT", "DELETED").Where("status = ?", status).Find(&analysisRequests)
				return analysisRequests, res.Error
			}

			res := r.db.Where("status != ? AND status != ?", "DRAFT", "DELETED").Where("status = ?", status).Where("formation_date < ?", endDate).
				Find(&analysisRequests)
			return analysisRequests, res.Error
		}

		if endDate.IsZero() {
			res := r.db.Where("status != ? AND status != ?", "DRAFT", "DELETED").Where("status = ?", status).Where("formation_date > ?", startDate).
				Find(&analysisRequests)
			return analysisRequests, res.Error
		}

		res := r.db.Where("status != ? AND status != ?", "DRAFT", "DELETED").Where("status = ?", status).Where("formation_date BETWEEN ? AND ?", startDate, endDate).
			Find(&analysisRequests)
		return analysisRequests, res.Error
	}

	if startDate.IsZero() {
		if endDate.IsZero() {
			// без фильтрации
			res := r.db.Where("status != ? AND status != ?", "DRAFT", "DELETED").Find(&analysisRequests)
			return analysisRequests, res.Error
		}

		// фильтрация по endDate
		res := r.db.Where("status != ? AND status != ?", "DRAFT", "DELETED").Where("formation_date < ?", endDate).
			Find(&analysisRequests)
		return analysisRequests, res.Error
	}

	if endDate.IsZero() {
		// фильтрация по startDate
		res := r.db.Where("status != ? AND status != ?", "DRAFT", "DELETED").Where("formation_date > ?", startDate).
			Find(&analysisRequests)
		return analysisRequests, res.Error
	}

	//фильтрация по startDate и endDate
	res := r.db.Where("status != ? AND status != ?", "DRAFT", "DELETED").Where("formation_date BETWEEN ? AND ?", startDate, endDate).
		Find(&analysisRequests)
	return analysisRequests, res.Error
}

func (r *Repo) GetAnalysisRequestById(requestId int) (models.AnalysisRequest, []models.Modeling, error) {
	var analysisRequest models.AnalysisRequest
	var modelings []models.Modeling

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

	return analysisRequest, modelings, nil
}

func (r *Repo) UpdateAnalysisRequestStatus(requestId int, status string) error {
	var analysisRequest models.AnalysisRequest
	err := r.db.First(&analysisRequest, "request_id = ?", requestId)
	if err.Error != nil {
		return err.Error
	}

	analysisRequest.Status = status
	if status == "REGISTERED" {
		analysisRequest.FormationDate = time.Now()
	}
	if status == "COMPLETE" {
		analysisRequest.CompleteDate = time.Now()
	}
	res := r.db.Save(&analysisRequest)

	return res.Error
}

func (r *Repo) DeleteModelingFromRequest(userId, modelingId int) (models.AnalysisRequest, []models.Modeling, error) {
	var request models.AnalysisRequest
	r.db.Where("user_id = ? and status = 'DRAFT'", userId).First(&request)

	if request.RequestId == 0 {
		return models.AnalysisRequest{}, nil, errors.New("no such request")
	}

	var requestModelings models.AnalysisRequestsModeling
	err := r.db.Where("request_id = ? AND modeling_id = ?", request.RequestId, modelingId).First(&requestModelings).Error
	if err != nil {
		return models.AnalysisRequest{}, nil, errors.New("такой услуги нет в данной заявке")
	}

	err = r.db.Where("request_id = ? AND modeling_id = ?", request.RequestId, modelingId).
		Delete(models.AnalysisRequestsModeling{}).Error

	if err != nil {
		return models.AnalysisRequest{}, nil, err
	}

	return r.GetAnalysisRequestById(request.RequestId)
}

func (r *Repo) UpdateModelingRequest(userId int, updateModelingRequest models.AnalysisRequestsModeling) error {
	var request models.AnalysisRequest
	r.db.Where("user_id = ? and status = 'DRAFT'", userId).First(&request)

	if request.RequestId == 0 {
		return errors.New("no such request")
	}

	var modelingRequest models.AnalysisRequestsModeling
	res := r.db.First(&modelingRequest, "modeling_id =? and request_id =?", updateModelingRequest.ModelingId, request.RequestId)
	if res.Error != nil {
		return res.Error
	}

	if updateModelingRequest.QueueSize != 0 {
		modelingRequest.QueueSize = updateModelingRequest.QueueSize
	}

	if updateModelingRequest.NodeQuantity != 0 {
		modelingRequest.NodeQuantity = updateModelingRequest.NodeQuantity
	}

	if updateModelingRequest.ClientQuantity != 0 {
		modelingRequest.ClientQuantity = updateModelingRequest.ClientQuantity
	}

	result := r.db.Save(modelingRequest)
	return result.Error
}
