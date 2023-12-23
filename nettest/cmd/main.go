package main

import (
	"context"
	"nettest/internal/pkg/app/handler"
	"nettest/internal/pkg/app/repo"
	"nettest/internal/pkg/db"
	"nettest/internal/pkg/minio"
)

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
