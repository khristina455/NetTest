package main

import (
	"context"
	"nettest/internal/pkg/app/handler"
	"nettest/internal/pkg/app/repo"
	"nettest/internal/pkg/db"
	"nettest/internal/pkg/minio"
	"nettest/internal/pkg/redis"
)

// @title AnalyzeNetworkApp
// @version 1.0
// @description App for analyze networks payload

// @host localhost:8080
// @schemes http
// @BasePath /
func main() {
	connection, _ := db.GetConnectionString()
	repo, _ := repo.NewRepository(connection)

	minioConfig := minio.InitConfig()

	minioClient, err := minio.NewMinioClient(context.Background(), minioConfig)
	if err != nil {
	}

	redisConfig := redis.InitRedisConfig()

	redisClient, err := redis.NewRedisClient(context.Background(), redisConfig)
	if err != nil {
	}

	handler := handler.NewHandler(repo, minioClient, redisClient)

	r := handler.InitRoutes()
	r.Run()
}
