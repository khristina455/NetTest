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
	err = db.AutoMigrate(&models.AnalysisRequest{})
	err = db.AutoMigrate(&models.AnalysisRequestsModeling{})
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

func (r *Repo) GetModelings(query string, from, to int) ([]models.Modeling, error) {
	modelings := make([]models.Modeling, 0)

	if query != "" {
		res := r.db.Where("is_deleted = ?", "false").Where("name LIKE ? AND price BETWEEN ? AND ?", "%"+query+"%", from, to).Find(&modelings)
		return modelings, res.Error
	}

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

func (r *Repo) GetUsersLoginForRequests(analysisRequests []models.AnalysisRequest) ([]models.AnalysisRequest, error) {
	for i := range analysisRequests {
		var user models.User
		r.db.Select("login").Where("user_id = ?", analysisRequests[i].UserId).First(&user)
		analysisRequests[i].User = user.Login
		fmt.Println(analysisRequests[i].User)

		r.db.Select("login").Where("user_id = ?", analysisRequests[i].AdminId).First(&user)
		analysisRequests[i].Admin = user.Login
		fmt.Println(analysisRequests[i].Admin)
	}
	return analysisRequests, nil
}

func (r *Repo) GetAnalysisRequests(status string, startDate, endDate time.Time, userId int, isAdmin bool) ([]models.AnalysisRequest, error) {
	var analysisRequests []models.AnalysisRequest
	ending := "AND creator_id = " + strconv.Itoa(userId)
	if isAdmin {
		ending = ""
	}

	if status != "" {
		if startDate.IsZero() {
			if endDate.IsZero() {
				res := r.db.Where("status != ? AND status != ?"+ending, "DRAFT", "DELETED").Where("status = ?", status).Find(&analysisRequests)
				analysisRequests, _ = r.GetUsersLoginForRequests(analysisRequests)
				return analysisRequests, res.Error
			}

			res := r.db.Where("status != ? AND status != ?"+ending, "DRAFT", "DELETED").Where("status = ?", status).Where("formation_date < ?", endDate).
				Find(&analysisRequests)
			analysisRequests, _ = r.GetUsersLoginForRequests(analysisRequests)
			return analysisRequests, res.Error
		}

		if endDate.IsZero() {
			res := r.db.Where("status != ? AND status != ?"+ending, "DRAFT", "DELETED").Where("status = ?", status).Where("formation_date > ?", startDate).
				Find(&analysisRequests)
			analysisRequests, _ = r.GetUsersLoginForRequests(analysisRequests)
			return analysisRequests, res.Error
		}

		res := r.db.Where("status != ? AND status != ?"+ending, "DRAFT", "DELETED").Where("status = ?", status).Where("formation_date BETWEEN ? AND ?", startDate, endDate).
			Find(&analysisRequests)
		analysisRequests, _ = r.GetUsersLoginForRequests(analysisRequests)
		return analysisRequests, res.Error
	}

	if startDate.IsZero() {
		if endDate.IsZero() {
			// без фильтрации
			res := r.db.Where("status != ? AND status != ?"+ending, "DRAFT", "DELETED").Find(&analysisRequests)
			analysisRequests, _ = r.GetUsersLoginForRequests(analysisRequests)
			return analysisRequests, res.Error
		}

		// фильтрация по endDate
		res := r.db.Where("status != ? AND status != ?"+ending, "DRAFT", "DELETED").Where("formation_date < ?", endDate).
			Find(&analysisRequests)
		analysisRequests, _ = r.GetUsersLoginForRequests(analysisRequests)
		return analysisRequests, res.Error
	}

	if endDate.IsZero() {
		// фильтрация по startDate
		res := r.db.Where("status != ? AND status != ?"+ending, "DRAFT", "DELETED").Where("formation_date > ?", startDate).
			Find(&analysisRequests)
		analysisRequests, _ = r.GetUsersLoginForRequests(analysisRequests)
		return analysisRequests, res.Error
	}

	//фильтрация по startDate и endDate
	res := r.db.Where("status != ? AND status != ?"+ending, "DRAFT", "DELETED").Where("formation_date BETWEEN ? AND ?", startDate, endDate).
		Find(&analysisRequests)
	analysisRequests, _ = r.GetUsersLoginForRequests(analysisRequests)
	return analysisRequests, res.Error
}

func (r *Repo) GetAnalysisRequestById(requestId int, userId int, isAdmin bool) (models.AnalysisRequest, []models.ModelingInRequestMessage, error) {
	var analysisRequest models.AnalysisRequest
	var modelings []models.Modeling
	var requestModeling []models.AnalysisRequestsModeling
	var modelingsWithFields []models.ModelingInRequestMessage

	result := r.db.First(&analysisRequest, "request_id =?", requestId)
	if result.Error != nil {
		return models.AnalysisRequest{}, nil, result.Error
	}

	if !isAdmin && analysisRequest.UserId != userId {
		return models.AnalysisRequest{}, nil, errors.New("ошибка доступа к данной заявке")
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

		modelingsWithFields = append(modelingsWithFields, currentRequest)
	}

	return analysisRequest, modelingsWithFields, nil
}

func (r *Repo) DeleteAnalysisRequest(userId int) error {
	var request models.AnalysisRequest
	res := r.db.First(&request, "user_id =? and status = 'DRAFT'", userId)
	if res.Error != nil {
		return res.Error
	}

	request.Status = "DELETED"
	result := r.db.Save(request)
	return result.Error
}

func (r *Repo) UpdateAnalysisRequestStatusClient(userId int, status string) error {
	var analysisRequest models.AnalysisRequest

	err := r.db.First(&analysisRequest, "user_id = ? AND status = ?", userId, "DRAFT")
	if err.Error != nil {
		return err.Error
	}

	analysisRequest.Status = status
	if status == "REGISTERED" {
		analysisRequest.FormationDate = time.Now()
	}

	res := r.db.Save(&analysisRequest)

	return res.Error
}

func (r *Repo) UpdateAnalysisRequestStatusAdmin(reqId int, status string) error {
	var analysisRequest models.AnalysisRequest

	err := r.db.First(&analysisRequest, "request_id = ? AND status = ?", reqId, "REGISTERED")
	if err.Error != nil {
		return err.Error
	}

	analysisRequest.Status = status
	if status == "COMPLETE" {
		analysisRequest.CompleteDate = time.Now()
	}

	res := r.db.Save(&analysisRequest)

	return res.Error
}

func (r *Repo) DeleteModelingFromRequest(userId, modelingId int) (models.AnalysisRequest, []models.ModelingInRequestMessage, error) {
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

	return r.GetAnalysisRequestById(request.RequestId, userId, false)
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

func (r *Repo) SignUp(newUser models.User) error {
	return r.db.Create(&newUser).Error
}

func (r *Repo) GetByCredentials(user models.User) (models.User, error) {
	err := r.db.First(&user, "login = ? AND password = ?", user.Login, user.Password).Error
	return user, err
}

func (r *Repo) GetUserInfo(user models.User) (models.User, error) {
	err := r.db.First(&user, "user_id = ?", user.UserId).Error
	return user, err
}
