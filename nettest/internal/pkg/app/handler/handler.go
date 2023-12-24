package handler

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"nettest/internal/models"
	"nettest/internal/pkg/app"
	"nettest/internal/pkg/minio"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	repo  app.Repo
	minio minio.Client
}

func NewHandler(repo app.Repo, minioClient minio.Client) *Handler {
	return &Handler{repo: repo, minio: minioClient}
}

func (h *Handler) InitRoutes() *gin.Engine {
	r := gin.Default()

	r.LoadHTMLGlob("templates/*")
	r.Static("/style", "./resources")

	r.GET("/modelings", h.GetModelingsList)
	r.GET("/modelings/:id", h.GetModeling)
	r.POST("/modelings", h.AddModeling)
	r.PUT("/modelings/:id", h.UpdateModeling)
	r.DELETE("/modelings/:id", h.DeleteModeling)
	r.POST("/modelings/request", h.AddModelingToRequest)

	r.GET("/analysis-requests", h.GetRequestsList)
	r.GET("/analysis-requests/:id", h.GetRequest)
	r.PUT("/analysis-requests/client", h.UpdateStatusClient)
	r.PUT("/analysis-requests/:id/admin", h.UpdateStatusAdmin)
	r.DELETE("/analysis-requests/:id", h.DeleteRequest)

	r.DELETE("/modelings/:id/requests", h.DeleteModelingFromRequest)
	r.PUT("/modelings/:id/requests", h.UpdateModelingRequest)

	r.Static("/images", "./resources")
	return r
}

// Взятие услуг
// Есть get параметры to from для фильтрации по цене
// Возвращает отфильтованные услуги и id черновой заявки
func (h *Handler) GetModelingsList(c *gin.Context) {
	to, _ := strconv.Atoi(c.Query("to"))
	from, _ := strconv.Atoi(c.Query("from"))

	if c.Query("to") == "" {
		to = 1e9
	}

	modelings, err := h.repo.GetModelings(from, to)
	if err != nil {
		log.Printf("cant get product by id %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	requestId, err := h.repo.GetDraftRequest(models.GetClientId())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"modelings": modelings, "draftId": requestId})
}

// Взятие услуги по id
func (h *Handler) GetModeling(c *gin.Context) {
	cardId := c.Param("id")
	id, err := strconv.Atoi(cardId)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	modeling, err := h.repo.GetModelingByID(id)
	if err != nil { // если не получилось
		log.Printf("cant get product by id %v", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"modeling": modeling})
}

// Добавление услуги
func (h *Handler) AddModeling(c *gin.Context) {
	var newModeling models.Modeling
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "ошибка при загрузке изображения"})
		return
	}

	newModeling.Name = c.Request.FormValue("name")
	if newModeling.Name == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "имя моделирования не может быть пустым"})
		return
	}

	newModeling.Description = c.Request.FormValue("description")

	price := c.Request.FormValue("price")
	fprice, err := strconv.ParseFloat(price, 32)
	if err != nil || fprice == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "цена указана неверно"})
		return
	}
	newModeling.Price = float32(fprice)

	if newModeling.Image, err = h.minio.SaveImage(c.Request.Context(), file, header); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "ошибка при сохранении изображения"})
		return
	}

	if err = h.repo.AddModeling(newModeling); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err})
		return
	}

	c.JSON(http.StatusCreated, "новая услуга успешно добавлена")
}

// Обновление услуги через id
func (h *Handler) UpdateModeling(c *gin.Context) {
	file, header, err := c.Request.FormFile("image")

	var updateModeling models.Modeling
	modelingId := c.Param("id")
	updateModeling.ModelingId, err = strconv.Atoi(modelingId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err})
	}

	updateModeling.Name = c.Request.FormValue("name")
	updateModeling.Description = c.Request.FormValue("description")

	price := c.Request.FormValue("price")
	if price != "" {
		fprice, err := strconv.ParseFloat(price, 32)
		if err != nil || fprice == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "цена указана неверно"})
			return
		}
		updateModeling.Price = float32(fprice)
	}
	if header != nil && header.Size != 0 {
		if updateModeling.Image, err = h.minio.SaveImage(c.Request.Context(), file, header); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err})
			return
		}

		url := h.repo.GetModelingImage(updateModeling.ModelingId)

		h.minio.DeleteImage(c.Request.Context(), strings.Split(url, "/")[4])
	}

	if err = h.repo.UpdateModeling(updateModeling); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "услуга успешно изменена"})
}

// Удаление услуги через id
func (h *Handler) DeleteModeling(c *gin.Context) {
	cardId := c.Param("id")
	id, err := strconv.Atoi(cardId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
	}

	err = h.repo.DeleteModelingByID(id)
	if err != nil { // если не получилось
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
	}
	c.JSON(http.StatusOK, "услуга удалена")
}

// Добавление услуги к завяки, в полях указывается ид услуги
func (h *Handler) AddModelingToRequest(c *gin.Context) {
	var request models.RequestCreateMessage

	err := c.BindJSON(&request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}
	request.UserId = models.GetClientId()

	err = h.repo.AddModelingToRequest(request)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "услуга добавлена в заявку"})
}

// Возвращение заявок пользователя отфильтрованных по дате и статусу,
// не должно быть черовиков и удаленных
func (h *Handler) GetRequestsList(c *gin.Context) {
	status := c.Query("status")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse(time.DateTime, startDateStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
	}

	if endDateStr != "" {
		endDate, err = time.Parse(time.DateTime, endDateStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}

		if endDate.Before(startDate) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "end_date не может быть раньше, чем start_date"})
			return
		}
	}

	analysisRequests, err := h.repo.GetAnalysisRequests(status, startDate, endDate)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, analysisRequests)
}

// Взятие заявки по ид и возвращает заявку с услугами и полями м-м
func (h *Handler) GetRequest(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}
	request, modelings, err := h.repo.GetAnalysisRequestById(id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"request": request, "modelings": modelings})
}

// Регистрация заявки-черновика клиентом
func (h *Handler) UpdateStatusClient(c *gin.Context) {
	var newRequestStatus models.AnalysisRequest
	err := c.BindJSON(&newRequestStatus)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	if newRequestStatus.Status != "REGISTERED" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Поменять статус можно только на 'REGISTERED'"})
		return
	}

	err = h.repo.UpdateAnalysisRequestStatus(0, newRequestStatus.Status)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Статус изменен"})
}

// Отмена/повреждение заявки админом по ид
func (h *Handler) UpdateStatusAdmin(c *gin.Context) {
	requestId := c.Param("id")
	id, err := strconv.Atoi(requestId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
	}

	var newRequestStatus models.AnalysisRequest
	err = c.BindJSON(&newRequestStatus)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	if newRequestStatus.Status != "COMPLETE" && newRequestStatus.Status != "CANCELED" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Поменять статус можно только на 'COMPLETE'"})
		return
	}

	err = h.repo.UpdateAnalysisRequestStatus(id, newRequestStatus.Status)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Статус изменен"})
}

func (h *Handler) DeleteRequest(c *gin.Context) {
	requestId := c.Param("id")
	id, err := strconv.Atoi(requestId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
	}

	err = h.repo.UpdateAnalysisRequestStatus(id, "DELETED")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Заявка удалена"})
}

func (h *Handler) DeleteModelingFromRequest(c *gin.Context) {
	modelingIdStr := c.Param("id")
	modelingId, err := strconv.Atoi(modelingIdStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	userId := models.GetClientId()

	request, modelings, err := h.repo.DeleteModelingFromRequest(userId, modelingId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Услуга удалена из заявки", "modelings": modelings, "request": request})
}

func (h *Handler) UpdateModelingRequest(c *gin.Context) {
	var updateModelingRequest models.AnalysisRequestsModeling
	var err error
	modelingId := c.Param("id")
	updateModelingRequest.ModelingId, err = strconv.Atoi(modelingId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err})
	}

	nodeQuantity := c.Request.FormValue("nodeQuantity")
	if nodeQuantity != "" {
		updateModelingRequest.NodeQuantity, err = strconv.Atoi(nodeQuantity)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err})
			return
		}
	}

	queueSize := c.Request.FormValue("queueSize")
	if queueSize != "" {
		updateModelingRequest.QueueSize, err = strconv.Atoi(queueSize)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err})
			return
		}
	}

	clientQuantity := c.Request.FormValue("clientQuantity")
	if clientQuantity != "" {
		updateModelingRequest.ClientQuantity, err = strconv.Atoi(clientQuantity)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err})
			return
		}
	}

	clientId := models.GetClientId()

	if err = h.repo.UpdateModelingRequest(clientId, updateModelingRequest); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "успешно изменено"})
}
