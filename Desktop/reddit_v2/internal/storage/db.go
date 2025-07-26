package storage

import (
	"context"
	"fmt"
	"reddit_v2/internal/models"

	"github.com/jackc/pgx/v5"
)

type Interface interface {
	Register(user *models.User) error
	Login(user *models.User) (models.User, error)
	GetAllPosts() ([]*models.Post, error)
	NewPost(post *models.Post) error
	GetPost(post_ID int) (*models.Post, error)
	GetPostsByCategory(category string) ([]*models.Post, error)
	GetPostsByUserLogin(username string) ([]*models.Post, error)
	GetUserName(authorID int) (string, error)
	AddComment(postID int, comment *models.Comment) (*models.Post, error)
	DeleteComment(idPost int, commentID int) (*models.Post, error)
	DeletePost(idPost int) ([]*models.Post, error)
	UpdateVote(idPost int, vote *models.Vote) (*models.Post, error)
}

type PostgresConnConfig struct {
	DBHost   string
	DBPort   uint
	DBName   string
	Username string
	Password string
	Options  map[string]string
}

type RedditDB struct {
	conn *pgx.Conn
}

func NewRedditDB(conn *pgx.Conn) *RedditDB {
	return &RedditDB{conn: conn}
}

func New(ctx context.Context, cfg PostgresConnConfig) (*pgx.Conn, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		cfg.Username, cfg.Password, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	var options string
	if len(cfg.Options) > 0 {
		for key, value := range cfg.Options {
			options += fmt.Sprintf("%s=%s&", key, value)
		}

		options = options[:len(options)-1]
		connStr += options
	}

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("error connecting to postgres: %w", err)
	}

	return conn, nil
}

func (s *RedditDB) Register(user *models.User) error {
	// Проверка существования пользователя
	var exists bool
	err := s.conn.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", user.Username).Scan(&exists)
	if err != nil {
		return fmt.Errorf("ошибка при проверке существования пользователя: %w", err)
	}

	// Если пользователь существует, возвращаем ошибку
	if exists {
		return fmt.Errorf("пользователь с именем %s уже существует", user.Username)
	}

	// Используем уже существующее соединение для вставки нового пользователя

	err = s.conn.QueryRow(context.Background(), "INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id", user.Username, user.Password).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("ошибка при вставке результата в таблицу: %w", err)
	}

	return nil
}

func (s *RedditDB) Login(user *models.User) (models.User, error) {
	var foundUser models.User
	err := s.conn.QueryRow(context.Background(), "SELECT id, username, password FROM users WHERE username = $1", user.Username).Scan(&foundUser.ID, &foundUser.Username, &foundUser.Password)
	if err != nil {
		return foundUser, fmt.Errorf("пользователь не найден: %w", err)
	}
	return foundUser, nil
}

func (s *RedditDB) GetAllPosts() ([]*models.Post, error) {
	rows, err := s.conn.Query(context.Background(), "SELECT id, title, url, author_id, category, score, created, views, type, text FROM Posts")
	if err != nil {
		return nil, fmt.Errorf("ошибка при поиске постов: %w", err)

	}
	defer rows.Close()

	var posts []*models.Post

	for rows.Next() {
		var post models.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.URL, &post.Author.ID, &post.Category, &post.Score, &post.Created, &post.Views, &post.Type, &post.Text); err != nil {
			return nil, fmt.Errorf("ошибка сканирования поста: ")
		}
		posts = append(posts, &post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}

	return posts, nil
}

func (s *RedditDB) NewPost(post *models.Post) error {
	// Выполняем запрос для вставки нового поста
	err := s.conn.QueryRow(
		context.Background(),
		"INSERT INTO Posts (title, url, author_id, category, score, type, text) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		post.Title,
		post.URL,
		post.Author.ID,
		post.Category,
		post.Score,
		post.Type,
		post.Text,
	).Scan(&post.ID)

	if err != nil {
		return fmt.Errorf("ошибка при вставке поста в таблицу: %w", err)
	}

	return nil
}

func (s *RedditDB) GetPost(post_ID int) (*models.Post, error) {
	var post models.Post

	// Увеличиваем количество просмотров на 1
	_, err := s.conn.Exec(context.Background(), `UPDATE Posts SET views = views + 1 WHERE id = $1`, post_ID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при увеличении количества просмотров: %w", err)
	}

	// Получение поста
	err = s.conn.QueryRow(context.Background(), `SELECT id, title, url, author_id, category, score, created, views, type, text FROM Posts WHERE id = $1`, post_ID).Scan(
		&post.ID,
		&post.Title,
		&post.URL,
		&post.Author.ID,
		&post.Category,
		&post.Score,
		&post.Created,
		&post.Views,
		&post.Type,
		&post.Text,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("пост с ID %d не найден", post_ID)
		}
		return nil, fmt.Errorf("ошибка при получении поста: %w", err)
	}

	// Получение комментариев для поста
	rows, err := s.conn.Query(context.Background(), `SELECT id, author_id, username, body, created FROM Comments WHERE post_id = $1`, post_ID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении комментариев: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var comment models.Comment
		if err := rows.Scan(&comment.ID, &comment.Author.ID, &comment.Author.Username, &comment.Body, &comment.Created); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании комментария: %w", err)
		}
		post.Comments = append(post.Comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при обработке результатов комментариев: %w", err)
	}

	return &post, nil
}

func (s *RedditDB) GetPostsByCategory(category string) ([]*models.Post, error) {
	// Выполняем SQL-запрос для получения постов по указанной категории
	rows, err := s.conn.Query(context.Background(), "SELECT id, title, url, author_id, category, score, created, views, type, text FROM Posts WHERE category = $1", category)
	if err != nil {
		return nil, fmt.Errorf("ошибка при поиске постов: %w", err)
	}
	defer rows.Close()

	var posts []*models.Post

	for rows.Next() {
		var post models.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.URL, &post.Author.ID, &post.Category, &post.Score, &post.Created, &post.Views, &post.Type, &post.Text); err != nil {
			return nil, fmt.Errorf("ошибка сканирования поста: %w", err)
		}
		posts = append(posts, &post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}

	return posts, nil
}

func (s *RedditDB) GetPostsByUserLogin(username string) ([]*models.Post, error) {
	// Сначала получаем ID автора по имени пользователя
	var authorID int
	err := s.conn.QueryRow(context.Background(), "SELECT id FROM Users WHERE username = $1", username).Scan(&authorID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при поиске пользователя: %w", err)
	}

	// Теперь получаем посты по ID автора
	rows, err := s.conn.Query(context.Background(), "SELECT id, title, url, author_id, category, score, created, views, type, text FROM Posts WHERE author_id = $1", authorID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при поиске постов: %w", err)
	}
	defer rows.Close()

	var posts []*models.Post

	for rows.Next() {
		var post models.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.URL, &post.Author.ID, &post.Category, &post.Score, &post.Created, &post.Views, &post.Type, &post.Text); err != nil {
			return nil, fmt.Errorf("ошибка сканирования поста: %w", err)
		}
		posts = append(posts, &post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}

	return posts, nil
}

func (s *RedditDB) GetUserName(authorID int) (string, error) {
	var userName string
	err := s.conn.QueryRow(context.Background(), "SELECT username FROM Users WHERE id = $1", authorID).Scan(&userName)
	if err != nil {
		return "", fmt.Errorf("ошибка при поиске пользователя: %w", err)
	}
	return userName, nil
}

func (s *RedditDB) AddComment(postID int, comment *models.Comment) (*models.Post, error) {
	// Вставка комментария в таблицу Comments
	var commentID int
	err := s.conn.QueryRow(context.Background(),
		"INSERT INTO Comments (author_id, post_id, username, body, created) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		comment.Author.ID, postID, comment.Author.Username, comment.Body, comment.Created).Scan(&commentID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при вставке комментария: %w", err)
	}
	post, err := s.GetPost(postID)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (s *RedditDB) DeleteComment(idPost int, commentID int) (*models.Post, error) {
	_, err := s.conn.Exec(context.Background(), `DELETE FROM Comments WHERE id = $1 AND post_id = $2`, commentID, idPost)
	if err != nil {
		return nil, fmt.Errorf("ошибка при удалении комментария: %w", err)
	}

	post, err := s.GetPost(idPost)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (s *RedditDB) DeletePost(idPost int) ([]*models.Post, error) {
	_, err := s.conn.Exec(context.Background(), `DELETE FROM Posts WHERE id = $1`, idPost)
	if err != nil {
		return nil, fmt.Errorf("ошибка при удалении комментария: %w", err)
	}

	posts, err := s.GetAllPosts()
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (s *RedditDB) UpdateVote(idPost int, vote *models.Vote) (*models.Post, error) {
	// Сначала получим пост, чтобы убедиться, что он существует
	post, err := s.GetPost(idPost)
	if err != nil {
		return nil, err
	}

	// Проверяем, существует ли голос
	var existingVote int
	queryCheck := `
		SELECT vote FROM Votes 
		WHERE user_id = $1 AND post_id = $2`
	err = s.conn.QueryRow(context.Background(), queryCheck, vote.User, idPost).Scan(&existingVote)

	if err != nil && err != pgx.ErrNoRows {
		// Если произошла ошибка, кроме ErrNoRows, возвращаем ее
		fmt.Println(err)
		return nil, fmt.Errorf("ошибка при проверке голоса: %w", err)
	}

	if err == pgx.ErrNoRows {
		// Если голоса не существует, вставляем новый
		queryInsert := `
			INSERT INTO Votes (user_id, post_id, vote)
			VALUES ($1, $2, $3)`
		_, err = s.conn.Exec(context.Background(), queryInsert, vote.User, idPost, vote.Vote)
		if err != nil {
			fmt.Println(err)
			return nil, fmt.Errorf("ошибка при вставке голоса: %w", err)
		}
	} else {
		// Если голос существует, обновляем его
		queryUpdate := `
			UPDATE Votes 
			SET vote = $1 
			WHERE user_id = $2 AND post_id = $3`
		_, err = s.conn.Exec(context.Background(), queryUpdate, vote.Vote, vote.User, idPost)
		if err != nil {
			fmt.Println(err)
			return nil, fmt.Errorf("ошибка при обновлении голоса: %w", err)
		}
	}
	postUpdate := `
			UPDATE Posts 
			SET score = score + $1
			WHERE id = $2 `
	_, err = s.conn.Exec(context.Background(), postUpdate, vote.Vote, idPost)
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("ошибка при обновлении голоса: %w", err)
	}
	return post, nil
}
