.PHONY: lint up

# TODO линтеры
lint:
	goloangci-lint run .

up:
	docker compose up
# запуск миграций
# go install github.com/pressly/goose/v3/cmd/goose@latest
# goose create init sql -dir migrations
PG_PASSWORD=qwerty
migrate:
	goose -dir migrations postgres "user=reddit_admin password=${PG_PASSWORD} dbname=reddit host=localhost port=5432 sslmode=disable" up

migrate-down:
	goose -dir migrations postgres "user=reddit_admin password=${PG_PASSWORD} dbname=reddit host=localhost port=5432 sslmode=disable" down
build:
	docker build -t web_app2:v1 .
run:
	docker run -d --name container1  -p 8080:8080 web_app2:v1
# docker compose