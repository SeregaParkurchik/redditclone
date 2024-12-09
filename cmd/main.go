package main

import (
	"fmt"
	"net/http"
	"redditclone/internal/core"
	"redditclone/internal/handler"
	"redditclone/internal/routes"
	"redditclone/internal/storage"
)

func main() {
	memStorage := storage.NewMemoryStorage()
	authService := core.New(memStorage)
	userHandler := handler.NewUserHandler(authService)

	mux := routes.InitRoutes(userHandler)
	fmt.Println("Запуск сервера на порту 8080 http://localhost:8080/")
	http.ListenAndServe(":8080", mux)
}
