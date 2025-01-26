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

# docker compose