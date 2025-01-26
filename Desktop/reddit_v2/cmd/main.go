package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"reddit_v2/internal/core"
	"reddit_v2/internal/handlers"
	"reddit_v2/internal/routes"
	"reddit_v2/internal/storage"
)

func main() {

	cfg := storage.PostgresConnConfig{
		DBHost:   "localhost",
		DBPort:   5432,
		DBName:   "reddit",
		Username: "reddit_admin",
		Password: "qwerty",
		Options:  nil, // или добавьте опции, если необходимо
	}

	// Создание соединения с базой данных
	conn, err := storage.New(context.Background(), cfg)
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer conn.Close(context.Background())

	// Создание экземпляра RedditDB
	redditDB := storage.NewRedditDB(conn)

	// Создание сервиса с использованием базы данных
	authService := core.New(redditDB)
	userHandler := handlers.NewUserHandler(authService)

	mux := routes.InitRoutes(userHandler)
	fmt.Println("Запуск сервера на порту 8080 http://localhost:8080/")
	http.ListenAndServe(":8080", mux)
}
