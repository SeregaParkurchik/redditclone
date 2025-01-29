-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS Users (
    id SERIAL PRIMARY KEY, -- Автоматически увеличивающийся идентификатор
    username VARCHAR(255) NOT NULL, -- Имя пользователя
    password VARCHAR(255) NOT NULL -- Пароль
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS Posts (
    id SERIAL PRIMARY KEY, -- Автоматически увеличивающийся идентификатор
    title VARCHAR(255) NOT NULL, -- Заголовок поста
    url VARCHAR(255) NOT NULL, -- URL поста
    author_id INT REFERENCES Users(id) ON DELETE CASCADE, -- Внешний ключ на таблицу Users
    category VARCHAR(255), -- Категория поста
    score INT DEFAULT 0, -- Оценка поста
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Дата создания поста
    views INT DEFAULT 0, -- Количество просмотров
    type VARCHAR(50), -- Тип поста
    text TEXT -- Текст поста
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS Votes (
    id SERIAL PRIMARY KEY, -- Автоматически увеличивающийся идентификатор
    user_id INT REFERENCES Users(id) ON DELETE CASCADE, -- Внешний ключ на таблицу Users
    post_id INT REFERENCES Posts(id) ON DELETE CASCADE, -- Внешний ключ на таблицу Posts
    vote INT NOT NULL -- Значение голоса
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS Comments (
    id SERIAL PRIMARY KEY, -- Автоматически увеличивающийся идентификатор
    author_id INT REFERENCES Users(id) ON DELETE CASCADE, -- Внешний ключ на таблицу Users
    post_id INT REFERENCES Posts(id) ON DELETE CASCADE, -- Внешний ключ на таблицу Posts
    username TEXT NOT NULL,
    body TEXT NOT NULL, -- Текст комментария
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP -- Дата создания комментария
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS Users;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS Posts;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS Votes;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS Comments;
-- +goose StatementEnd
