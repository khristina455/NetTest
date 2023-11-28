package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"nettest/internal/pkg/app"
	"strconv"
)

type Handler struct {
	repo app.Repo
}

func NewHandler(repo app.Repo) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) InitRoutes() *gin.Engine {
	r := gin.Default()

	r.LoadHTMLGlob("templates/*")
	r.Static("/style", "./resources")

	r.GET("/", h.GetCardsList)
	r.GET("/products/:id", h.GetCardById)
	r.POST("/products/:id", h.DeleteCard)

	r.Static("/images", "./resources")
	return r
}

func (h *Handler) DeleteCard(c *gin.Context) {
	cardId := c.Param("id")
	id, err := strconv.Atoi(cardId)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	err = h.repo.DeleteModelingByID(id)
	if err != nil { // если не получилось
		log.Printf("cant get product by id %v", err)
		c.Error(err)
		return
	}
	c.Redirect(http.StatusFound, "/")
}

func (h *Handler) GetCardsList(c *gin.Context) {
	to, _ := strconv.Atoi(c.Query("to"))
	from, _ := strconv.Atoi(c.Query("from"))

	if c.Query("to") == "" {
		to = 1e9
	}

	modelings, err := h.repo.GetModelings(from, to)
	if err != nil {
		log.Printf("cant get product by id %v", err)
		c.Error(err)
		return
	}

	fmt.Println(modelings)
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title":    "Nettest",
		"products": modelings,
	})
}

func (h *Handler) GetCardById(c *gin.Context) {
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

	c.HTML(http.StatusOK, "card.html", gin.H{
		"title":   "Nettest",
		"product": modeling,
	})
}
