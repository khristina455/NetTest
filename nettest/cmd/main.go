package main

import (
	"context"
	"nettest/internal/pkg/app/handler"
	"nettest/internal/pkg/app/repo"
	"nettest/internal/pkg/db"
	"nettest/internal/pkg/minio"
)

// @title BITOP
// @version 1.0
// @description Bmstu Open IT Platform

// @contact.name API Support
// @contact.url https://vk.com/bmstu_schedule
// @contact.email bitop@spatecon.ru

// @license.name AS IS (NO WARRANTY)

// @host 127.0.0.1
// @schemes https http
// @BasePath /
func main() {
	connection, _ := db.GetConnectionString()
	repo, _ := repo.NewRepository(connection)

	minioConfig := minio.InitConfig()

	minioClient, err := minio.NewMinioClient(context.Background(), minioConfig)
	if err != nil {
	}

	handler := handler.NewHandler(repo, minioClient)

	r := handler.InitRoutes()
	r.Run()
}
