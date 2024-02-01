package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
	"net/http"
	_ "nettest/docs"
	"nettest/internal/models"
	"nettest/internal/pkg/app"
	"nettest/internal/pkg/auth"
	"nettest/internal/pkg/minio"
	"nettest/internal/pkg/redis"
	"os"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	repo         app.Repo
	minio        minio.Client
	redis        redis.Client
	tokenManager auth.TokenManager
	hasher       auth.PasswordHasher
}

func NewHandler(repo app.Repo, minioClient minio.Client, client redis.Client) *Handler {
	tokenManager, err := auth.NewManager(os.Getenv("TOKEN_SECRET"))
	if err != nil {
	}
	return &Handler{repo: repo, minio: minioClient, redis: client, tokenManager: tokenManager}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", c.GetHeader("Origin"))
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func (h *Handler) InitRoutes() *gin.Engine {
	r := gin.Default()
	r.Use(CORSMiddleware())

	r.LoadHTMLGlob("templates/*")
	r.Static("/style", "./resources")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiGroup := r.Group("/api")
	{
		apiGroup.GET("/modelings", h.WithAuthCheck([]models.Role{}), h.GetModelingsList)
		apiGroup.GET("/modelings/:id", h.GetModeling)
		apiGroup.POST("/modelings", h.WithAuthCheck([]models.Role{models.Admin}), h.AddModeling)
		apiGroup.PUT("/modelings/:id", h.WithAuthCheck([]models.Role{models.Admin}), h.UpdateModeling)
		apiGroup.DELETE("/modelings/:id", h.WithAuthCheck([]models.Role{models.Admin}), h.DeleteModeling)
		apiGroup.POST("/modelings/request", h.WithAuthCheck([]models.Role{models.Client}), h.AddModelingToRequest)

		apiGroup.GET("/analysis-requests", h.WithAuthCheck([]models.Role{models.Admin, models.Client}), h.GetRequestsList)
		apiGroup.GET("/analysis-requests/:id", h.WithAuthCheck([]models.Role{models.Admin, models.Client}), h.GetRequest)
		apiGroup.PUT("/analysis-requests/client", h.WithAuthCheck([]models.Role{models.Client}), h.UpdateStatusClient)
		apiGroup.PUT("/analysis-requests/:id/admin", h.WithAuthCheck([]models.Role{models.Admin}), h.UpdateStatusAdmin)
		apiGroup.DELETE("/analysis-requests", h.WithAuthCheck([]models.Role{models.Client}), h.DeleteRequest)

		apiGroup.DELETE("/analysis-requests/modelings/:id", h.WithAuthCheck([]models.Role{models.Client}), h.DeleteModelingFromRequest)
		apiGroup.PUT("/analysis-requests/modelings/:id", h.WithAuthCheck([]models.Role{models.Client}), h.UpdateModelingRequest)

		apiGroup.POST("/signIn", h.SignIn)
		apiGroup.POST("/signUp", h.SignUp)
		apiGroup.POST("/logout", h.Logout)
		apiGroup.GET("/checkAuth", h.WithAuthCheck([]models.Role{models.Client, models.Admin}), h.CheckAuth)
	}

	r.Static("/images", "./resources")
	return r
}

// Взятие услуг
// Есть get параметры to from для фильтрации по цене
// Возвращает отфильтованные услуги и id черновой заявки

// GetModelingsList godoc
// @Summary      Get list of modelings
// @Description  Retrieves a list of modelings based on the provided parameters
// @Tags         Modelings
// @Accept       json
// @Produce      json
// @Param        query   query    string  false  "Query string to filter modelings"
// @Success      200  {object}  map[string]any
// @Failure      500  {object}  error
// @Router       /api/modelings [get]
func (h *Handler) GetModelingsList(c *gin.Context) {
	query := c.Query("query")
	to, _ := strconv.Atoi(c.Query("to"))
	from, _ := strconv.Atoi(c.Query("from"))

	if c.Query("to") == "" {
		to = 1e9
	}

	modelings, err := h.repo.GetModelings(query, from, to)
	if err != nil {
		log.Printf("cant get product by id %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	requestId, err := h.repo.GetDraftRequest(c.GetInt(userCtx))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"modelings": modelings, "draftId": requestId})
}

// Взятие услуги по id

// GetModeling godoc
// @Summary      Get modeling by ID
// @Description  Retrieves a modeling by its ID
// @Tags         Modelings
// @Produce      json
// @Param        id   path    int     true        "Modeling ID"
// @Success      200  {object}  models.Modeling
// @Failure      400  {object}  error
// @Router       /api/modelings/{id} [get]
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

	c.JSON(http.StatusOK, modeling)
}

// Добавление услуги

// AddModeling godoc
// @Summary      Add new modeling
// @Description  Add a new modeling with image, name, description, and price
// @Tags         Modelings
// @Accept       multipart/form-data
// @Produce      json
// @Param        image formData file true "Modeling image"
// @Param        name formData string true "Modeling name"
// @Param        description formData string false "Modeling description"
// @Param        price formData integer true "Modeling price"
// @Success      201  {string}  map[string]any
// @Failure      400  {object}  map[string]any
// @Failure      500  {object}  map[string]any
// @Router       /api/modelings [post]
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

// UpdateModeling godoc
// @Summary      Update modeling by ID
// @Description  Updates a modeling with the given ID
// @Tags         Modelings
// @Accept       multipart/form-data
// @Produce      json
// @Param        id          path        int     true        "ID"
// @Param        name        formData    string  false       "name"
// @Param        description formData    string  false       "description"
// @Param        price       formData    string  false       "price"
// @Param        image       formData    file    false       "image"
// @Success      200         {object}    map[string]any
// @Failure      400         {object}    error
// @Router       /api/modelings/{id} [put]
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

// DeleteModeling godoc
// @Summary      Delete modeling by ID
// @Description  Deletes a modeling with the given ID
// @Tags         Modelings
// @Accept       json
// @Produce      json
// @Param        id  path  int  true  "Modeling ID"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  error
// @Router       /api/modelings/{id} [delete]
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

// AddModelingToRequest godoc
// @Summary      Add modeling to request
// @Description  Adds a modeling to analysis request
// @Tags         Modelings
// @Accept       json
// @Produce      json
// @Param        modelingId  path  int  true  "Modeling ID"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  error
// @Router       /api/modelings/request [post]
func (h *Handler) AddModelingToRequest(c *gin.Context) {
	var request models.RequestCreateMessage
	request.UserId = c.GetInt(userCtx)

	err := c.BindJSON(&request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	err = h.repo.AddModelingToRequest(request)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "услуга добавлена в заявку"})
}

// Возвращение заявок пользователя отфильтрованных по дате и статусу,
// не должно быть черовиков и удаленных

// GetRequestsList godoc
// @Summary      Get list of analysis requests
// @Description  Retrieves a list of analysis requests based on the provided parameters
// @Tags         AnalysisRequests
// @Accept       json
// @Produce      json
// @Param        status      query  string    false  "Analysis request status"
// @Param        start_date  query  string    false  "Start date in the format '2006-01-02T15:04:05Z'"
// @Param        end_date    query  string    false  "End date in the format '2006-01-02T15:04:05Z'"
// @Success      200  {object}  []models.AnalysisRequest
// @Failure      400  {object}  error
// @Failure      500  {object}  error
// @Router       /api/analysis-requests [get]
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

	analysisRequests, err := h.repo.GetAnalysisRequests(status, startDate, endDate, c.GetInt(userCtx), c.GetBool(adminCtx))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, analysisRequests)
}

// Взятие заявки по ид и возвращает заявку с услугами и полями м-м

// GetRequest godoc
// @Summary      Get analysis request by ID
// @Description  Retrieves an analysis request with the given ID
// @Tags         AnalysisRequests
// @Accept       json
// @Produce      json
// @Param        id  path  int  true  "Analysis Request ID"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  error
// @Router       /api/analysis-requests/{id} [get]
func (h *Handler) GetRequest(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}
	request, modelings, err := h.repo.GetAnalysisRequestById(id, c.GetInt(userCtx), c.GetBool(adminCtx))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"request": request, "modelings": modelings})
}

// Регистрация заявки-черновика клиентом

// UpdateStatusClient godoc
// @Summary      Update analysis request status by client
// @Description  Updates the status of an analysis request by client on registered
// @Tags         AnalysisRequests
// @Accept       json
// @Produce      json
// @Param        status    body    models.AnalysisRequest  true    "New status of the analysis request"
// @Success      200          {object}  map[string]string
// @Failure      400          {object}  error
// @Router       /api/analysis-requests/client [put]
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

	err = h.repo.UpdateAnalysisRequestStatusClient(c.GetInt(userCtx), newRequestStatus.Status)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Статус изменен"})
}

// Отмена/повреждение заявки админом по ид

// UpdateStatusAdmin godoc
// @Summary      Update analysis request status by ID
// @Description  Updates the status of an analysis request with the given ID on "COMPLETE"/"CANCELED"
// @Tags         AnalysisRequests
// @Accept       json
// @Produce      json
// @Param        requestId  path  int  true  "Request ID"
// @Param        status  body  models.AnalysisRequest  true  "New request status"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  error
// @Router       /api/analysis-requests/{requestId}/admin [put]
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

	err = h.repo.UpdateAnalysisRequestStatusAdmin(id, newRequestStatus.Status)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Статус изменен"})
}

// DeleteRequest godoc
// @Summary      Delete analysis request by user ID
// @Description  Deletes an analysis request for the given user ID
// @Tags         AnalysisRequests
// @Accept       json
// @Produce      json
// @Param        user_id  path  int  true  "User ID"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  error
// @Router       /api/analysis-requests/{requestId} [delete]
func (h *Handler) DeleteRequest(c *gin.Context) {
	userId := c.GetInt(userCtx)
	err := h.repo.DeleteAnalysisRequest(userId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Заявка удалена"})
}

// DeleteModelingFromRequest godoc
// @Summary      Delete modeling from request
// @Description  Deletes a modeling from a request based on the user ID and threat ID
// @Tags         AnalysisRequests
// @Accept       json
// @Produce      json
// @Param        modelingId  path  int  true  "Modeling ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  error
// @Router       /api/modelings/{modelingId}/requests [delete]
func (h *Handler) DeleteModelingFromRequest(c *gin.Context) {
	modelingIdStr := c.Param("id")
	modelingId, err := strconv.Atoi(modelingIdStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	userId := c.GetInt(userCtx)

	request, modelings, err := h.repo.DeleteModelingFromRequest(userId, modelingId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Услуга удалена из заявки", "modelings": modelings, "request": request})
}

// UpdateModelingRequest godoc
// @Summary      Update request_modeling by ID
// @Description  Updates a request_modeling the given ID
// @Tags         RequestsModelings
// @Accept       multipart/form-data
// @Produce      json
// @Param        id          path        int     true        "ID"
// @Param        nodeQuantity        formData    string  false       "nodeQuantity"
// @Param        queueSize           formData    string  false       "queueSize"
// @Param        clientQuantity      formData    string  false       "clientQuantity"
// @Success      200         {object}    map[string]any
// @Failure      400         {object}    error
// @Router       /api/modelings/{modelingId}/requests [put]
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

	clientId := c.GetInt(userCtx)

	if err = h.repo.UpdateModelingRequest(clientId, updateModelingRequest); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "успешно изменено"})
}

// SignIn godoc
// @Summary      User sign-in
// @Description  Authenticates a user and generates an access token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        user  body  models.UserLogin  true  "User information"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  error
// @Failure      401  {object}  error
// @Failure      500  {object}  error
// @Router       /api/signIn [post]
func (h *Handler) SignIn(c *gin.Context) {
	var clientInfo models.UserLogin
	var err error

	if err = c.BindJSON(&clientInfo); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, "неверный формат данных")
		return
	}

	user, err := h.repo.GetByCredentials(models.User{Password: clientInfo.Password, Login: clientInfo.Login})
	if err != nil {
		fmt.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "ошибка авторизации"})
		return
	}

	token, err := h.tokenManager.NewJWT(user.UserId, user.IsAdmin)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "ошибка при формировании токена"})
		return
	}

	c.SetCookie("AccessToken", "Bearer "+token, 0, "/", "localhost", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "клиент успешно авторизован", "isAdmin": user.IsAdmin, "login": user.Login, "userId": user.UserId})
}

// SignUp godoc
// @Summary      Sign up a new user
// @Description  Creates a new user account
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        user  body  models.UserSignUp  true  "User information"
// @Success      201  {object}  map[string]any
// @Failure      400  {object}  error
// @Failure      409  {object}  error
// @Failure      500  {object}  error
// @Router       /api/signUp [post]
func (h *Handler) SignUp(c *gin.Context) {
	var newClient models.UserSignUp
	var err error

	if err = c.BindJSON(&newClient); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "неверный формат данных о новом пользователе"})
		return
	}

	if err = h.repo.SignUp(models.User{
		Login:    newClient.Login,
		Password: newClient.Password,
	}); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "нельзя создать пользователя с таким логином"})

		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "пользователь успешно создан"})
}

// Logout godoc
// @Summary      Logout
// @Description  Logs out the user by blacklisting the access token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Success      200
// @Failure      400
// @Router       /api/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	jwtStr, err := c.Cookie("AccessToken")
	if !strings.HasPrefix(jwtStr, jwtPrefix) || err != nil { // если нет префикса то нас дурят!
		c.AbortWithStatus(http.StatusBadRequest) // отдаем что нет доступа
		return
	}

	// отрезаем префикс
	jwtStr = jwtStr[len(jwtPrefix):]

	_, _, err = h.tokenManager.Parse(jwtStr)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		log.Println(err)
		return
	}

	// сохраняем в блеклист редиса
	err = h.redis.WriteJWTToBlacklist(c.Request.Context(), jwtStr, time.Hour)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}

// CheckAuth godoc
// @Summary      Check user authentication
// @Description  Retrieves user information based on the provided user context
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Success      200  {object}  models.User
// @Failure      500  {object}  string
// @Router       /api/check-auth [get]
func (h *Handler) CheckAuth(c *gin.Context) {
	var userInfo = models.User{UserId: c.GetInt(userCtx)}

	userInfo, err := h.repo.GetUserInfo(userInfo)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, "неверный формат данных")
		return
	}

	c.JSON(http.StatusOK, userInfo)
}
