package models

import "time"

type User struct {
	ID       int    `json:"id" db:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Post struct {
	ID       int       `json:"id"`
	Title    string    `json:"title"`    // Заголовок поста
	URL      string    `json:"url"`      // URL поста
	Author   User      `json:"author"`   // ID автора
	Category string    `json:"category"` // Категория поста
	Score    int       `json:"score"`    // Оценка поста
	Votes    []Vote    `json:"votes"`    // Список голосов
	Comments []Comment `json:"comments"` // Список комментариев
	Created  time.Time `json:"created"`  // Дата создания поста
	Views    int       `json:"views"`    // Количество просмотров
	Type     string    `json:"type"`     // Тип поста
	Text     string    `json:"text"`     // Текст поста
}

type Vote struct {
	User int `json:"user"` // ID пользователя
	Vote int `json:"vote"` // Значение голоса
}

type Comment struct {
	ID      int       `json:"id"`
	Author  User      `json:"author"`
	Body    string    `json:"body"`    // Текст комментария
	Created time.Time `json:"created"` // Дата создания комментария
}
