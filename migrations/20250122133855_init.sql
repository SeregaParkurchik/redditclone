-- +goose Up
-- Создание всех таблиц
CREATE TABLE IF NOT EXISTS Users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS Posts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    url VARCHAR(255) NOT NULL,
    author_id INT REFERENCES Users(id) ON DELETE CASCADE,
    category VARCHAR(255),
    score INT DEFAULT 0,
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    views INT DEFAULT 0,
    type VARCHAR(50),
    text TEXT
);

CREATE TABLE IF NOT EXISTS Votes (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES Users(id) ON DELETE CASCADE,
    post_id INT REFERENCES Posts(id) ON DELETE CASCADE,
    vote INT NOT NULL
);

CREATE TABLE IF NOT EXISTS Comments (
    id SERIAL PRIMARY KEY,
    author_id INT REFERENCES Users(id) ON DELETE CASCADE,
    post_id INT REFERENCES Posts(id) ON DELETE CASCADE,
    username TEXT NOT NULL,
    body TEXT NOT NULL,
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Теперь, когда все таблицы созданы, добавляем ограничение
-- Оно необходимо для корректной работы 'ON CONFLICT (user_id, post_id)'
ALTER TABLE Votes ADD CONSTRAINT unique_user_post UNIQUE (user_id, post_id);


-- +goose Down
-- Откат изменений в обратном порядке

-- Сначала удаляем ограничение
ALTER TABLE Votes DROP CONSTRAINT IF EXISTS unique_user_post;

-- Затем удаляем таблицы
DROP TABLE IF EXISTS Comments;
DROP TABLE IF EXISTS Votes;
DROP TABLE IF EXISTS Posts;
DROP TABLE IF EXISTS Users;
