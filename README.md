# reddit_clone

Этот проект представляет собой учебный клон Reddit, написанный на Go, с готовым фронтендом.

---

### Технологии

- **Бэкенд:**
  - **Go 1.22.9**
  - **Роутер:** [gorilla/mux](https://github.com/gorilla/mux)
  - **База данных:** PostgreSQL
  - **Драйвер БД:** [jackc/pgx/v5](https://github.com/jackc/pgx)
  - **Аутентификация:** JWT ([golang-jwt/jwt](https://github.com/golang-jwt/jwt))
  - **Хэширование паролей:** [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto)

---

### Требования

- **Docker** и **Docker Compose** для запуска базы данных.
- **Go 1.22.9** или выше.
- **`make`** для использования скриптов сборки и запуска.

---

### Запуск проекта
1.  **Запуск базы данных:**
    ```bash
    make up
    ```
    Эта команда запускает контейнер с PostgreSQL.

2.  **Выполнение миграций:**
    ```bash
    make migrate
    ```
    Эта команда применяет миграции из папки `migrations` к базе данных.

3.  **Сборка проекта:**
    ```bash
    make build
    ```
    Эта команда компилирует Go-приложение.

4.  **Запуск приложения:**
    ```bash
    make run
    ```
    После запуска приложение будет доступно по адресу `http://localhost:8080`.


