package main

import (
	"nettest/internal/pkg/app/handler"
	"nettest/internal/pkg/app/repo"
	"nettest/internal/pkg/db"
)

func main() {
	connection, _ := db.GetConnectionString()
	repo, _ := repo.NewRepository(connection)

	handler := handler.NewHandler(repo)

	r := handler.InitRoutes()
	r.Run()
}
