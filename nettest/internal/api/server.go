package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Product struct {
	Id          int
	Name        string
	Description string
	ImgSrc      string
	Price       int
}

var products = []Product{
	Product{1, "Аналитическое моделирование очереди в узле сети", "Рассчет времени ожидания в узле сети", "/images/card1.jpg", 1500},
	Product{2, "Аналитическое моедлирование прохождения сообщения в сети", "Рассчет среденего времени прохождения сообщения в сети", "/images/card2.jpeg", 5000},
	Product{3, "Симуляционное моделирование очередей", "Сбор статистика о таких величинах: cредняя длина очереди, пиковая длина очереди, среднеквадратичное отклонение длины очереди от среднего значения", "/images/card3.png", 10000},
	Product{4, "Симуляционное моделирование времени ожидания", "Сбор статистика о таких величинах: cреднее время ожидания, максимальное время ожидания, среднеквадратичное отклонение времени ожидания", "/images/card4.png", 10000},
	Product{5, "Симуляционное моделирование системного времени", "Сбор статистика о таких величинах: среднее системное время, максимальное системное время, среднеквадратичное отклонение системного времени, полное число сообщений в статистике системного времени, пиковое значение числа системных сообщений, среднеквадратичное отклонение числа системных сообщений", "/images/card5.jpeg", 15000},
	Product{6, "Симуляционное моделирование потерь сообщения", "Сбор статистика о таких величинах: полное число потерянных сообщений, частота потери сообщений, доля потерь из-за переполнения очереди, доля потерь из-за таймаутов", "/images/card6.png", 12500},
}

func StartServer() {
	log.Println("Server start up")

	r := gin.Default()

	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) {
		to, _ := strconv.Atoi(c.Query("to"))
		from, _ := strconv.Atoi(c.Query("from"))

		if c.Query("to") == "" {
			to = 1e9
		}

		var filterProducts []Product

		for _, product := range products {
			if product.Price >= from && product.Price <= to {
				filterProducts = append(filterProducts, product)
			}
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"title":    "NetTest",
			"products": filterProducts,
			"to":       c.Query("to"),
			"from":     c.Query("from"),
		})
	})

	r.GET("/products/:id", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		c.HTML(http.StatusOK, "card.html", gin.H{
			"product": products[id-1],
		})
	})

	r.Static("/style", "./resources/")
	r.Static("/images", "./resources/")

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

	log.Println("Server down")
}
