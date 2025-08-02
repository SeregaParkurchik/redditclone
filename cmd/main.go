package main

import (
	"context"
	"fmt"
	"log"
	"log/slog" // Импортируем slog для логирования
	"net/http"
	"os" // Импортируем os для работы с стандартным выводом

	"reddit_v2/internal/core"
	"reddit_v2/internal/handlers"
	"reddit_v2/internal/pg" // Импортируем нашу обертку
	"reddit_v2/internal/routes"
	"reddit_v2/internal/storage"
)

func main() {
	// 1. Инициализация логгера
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// 2. Определение строки подключения
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		"reddit_admin", "qwerty", "host.docker.internal", 5432, "reddit",
	)

	// 3. Инициализация нашей обертки, которая создает пул соединений
	dbClient, err := pg.NewDB(context.Background(), connString, logger)
	if err != nil {
		log.Fatalf("Не удалось инициализировать обертку базы данных: %v", err)
	}
	defer dbClient.Close() // Закрываем пул при завершении работы приложения

	// 4. Создание экземпляра хранилища с использованием нашей обертки
	redditDB := storage.NewRedditDB(dbClient)

	// 5. Создание сервиса и обработчиков
	authService := core.New(redditDB)
	userHandler := handlers.NewUserHandler(authService)

	// 6. Запуск сервера
	mux := routes.InitRoutes(userHandler)
	fmt.Println("Запуск сервера на порту 8080 http://localhost:8080/")
	http.ListenAndServe(":8080", mux)
}
