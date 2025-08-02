package storage

import (
	"context"
	"errors"
	"fmt"
	"reddit_v2/internal/models"
	"reddit_v2/internal/pg"

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
	Close()
}

type RedditDB struct {
	db *pg.DB
}

func NewRedditDB(db *pg.DB) *RedditDB {
	return &RedditDB{db: db}
}

func (s *RedditDB) Close() {
	s.db.Close()
}

func (s *RedditDB) Register(user *models.User) error {
	ctx := context.Background()

	var exists bool
	sql := "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)"
	err := s.db.QueryOne(ctx, &exists, sql, user.Username)

	if err != nil {
		return fmt.Errorf("ошибка при проверке существования пользователя: %w", err)
	}

	if exists {
		return fmt.Errorf("пользователь с именем %s уже существует", user.Username)
	}

	sql = "INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id"
	err = s.db.QueryOne(ctx, &user.ID, sql, user.Username, user.Password)
	if err != nil {
		return fmt.Errorf("ошибка при вставке нового пользователя: %w", err)
	}

	return nil
}

func (s *RedditDB) Login(user *models.User) (models.User, error) {
	var foundUser models.User

	query := "SELECT id, username, password FROM users WHERE username = $1"
	err := s.db.QueryOne(context.Background(), &foundUser, query, user.Username)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return foundUser, fmt.Errorf("пользователь не найден: %w", err)
		}
		return foundUser, fmt.Errorf("ошибка при поиске пользователя: %w", err)
	}

	return foundUser, nil
}

func (s *RedditDB) GetAllPosts() ([]*models.Post, error) {
	var posts []*models.Post

	query := `
        SELECT
            p.id, p.title, p.url, p.category, p.score, p.created, p.views, p.type, p.text,
            u.id AS "author.id",
            u.username AS "author.username"
        FROM Posts p
        JOIN Users u ON u.id = p.author_id`

	err := s.db.QueryMany(context.Background(), &posts, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка при поиске постов: %w", err)
	}

	return posts, nil
}

func (s *RedditDB) NewPost(post *models.Post) error {
	query := `
        INSERT INTO Posts (title, url, author_id, category, score, type, text)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id`
	err := s.db.QueryOne(
		context.Background(),
		&post.ID,
		query,
		post.Title,
		post.URL,
		post.Author.ID,
		post.Category,
		post.Score,
		post.Type,
		post.Text,
	)

	if err != nil {
		return fmt.Errorf("ошибка при вставке поста в таблицу: %w", err)
	}

	return nil
}

func (s *RedditDB) GetPost(post_ID int) (*models.Post, error) {
	ctx := context.Background()
	var post models.Post

	_, err := s.db.Exec(ctx, `UPDATE Posts SET views = views + 1 WHERE id = $1`, post_ID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при увеличении количества просмотров: %w", err)
	}

	queryPost := `
        SELECT
            p.id, p.title, p.url, p.category, p.score, p.created, p.views, p.type, p.text,
            u.id AS "author.id",
            u.username AS "author.username"
        FROM Posts p
        JOIN Users u ON u.id = p.author_id
        WHERE p.id = $1`
	err = s.db.QueryOne(ctx, &post, queryPost, post_ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("пост с ID %d не найден", post_ID)
		}
		return nil, fmt.Errorf("ошибка при получении поста: %w", err)
	}

	var comments []models.Comment

	queryComments := `
        SELECT
            c.id, c.body, c.created,
            u.id AS "author.id",
            u.username AS "author.username"
        FROM Comments c
        JOIN Users u ON u.id = c.author_id
        WHERE c.post_id = $1`
	err = s.db.QueryMany(ctx, &comments, queryComments, post_ID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении комментариев: %w", err)
	}
	post.Comments = comments

	return &post, nil
}

func (s *RedditDB) GetPostsByCategory(category string) ([]*models.Post, error) {
	var posts []*models.Post
	query := `
        SELECT
            p.id, p.title, p.url, p.category, p.score, p.created, p.views, p.type, p.text,
            u.id AS "author.id",
            u.username AS "author.username"
        FROM Posts p
        JOIN Users u ON u.id = p.author_id
        WHERE p.category = $1`
	err := s.db.QueryMany(context.Background(), &posts, query, category)
	if err != nil {
		return nil, fmt.Errorf("ошибка при поиске постов: %w", err)
	}

	return posts, nil
}

func (s *RedditDB) GetPostsByUserLogin(username string) ([]*models.Post, error) {
	var posts []*models.Post

	query := `
        SELECT
            p.id, p.title, p.url, p.category, p.score, p.created, p.views, p.type, p.text,
            u.id AS "author.id",
            u.username AS "author.username"
        FROM Posts p
        JOIN Users u ON u.id = p.author_id
        WHERE u.username = $1`
	err := s.db.QueryMany(context.Background(), &posts, query, username)
	if err != nil {
		return nil, fmt.Errorf("ошибка при поиске постов по имени пользователя: %w", err)
	}

	return posts, nil
}

func (s *RedditDB) GetUserName(authorID int) (string, error) {
	var userName string
	query := `SELECT username FROM Users WHERE id = $1`
	err := s.db.QueryOne(context.Background(), &userName, query, authorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("пользователь с ID %d не найден", authorID)
		}
		return "", fmt.Errorf("ошибка при поиске пользователя: %w", err)
	}
	return userName, nil
}

func (s *RedditDB) AddComment(postID int, comment *models.Comment) (*models.Post, error) {
	ctx := context.Background()
	var commentID int
	query := `
        INSERT INTO Comments (author_id, post_id, username, body, created)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id`
	err := s.db.QueryOne(
		ctx,
		&commentID,
		query,
		comment.Author.ID,
		postID,
		comment.Author.Username,
		comment.Body,
		comment.Created,
	)

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
	_, err := s.db.Exec(context.Background(), `DELETE FROM Comments WHERE id = $1 AND post_id = $2`, commentID, idPost)
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
	_, err := s.db.Exec(context.Background(), `DELETE FROM Posts WHERE id = $1`, idPost)
	if err != nil {
		return nil, fmt.Errorf("ошибка при удалении поста: %w", err)
	}

	posts, err := s.GetAllPosts()
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (s *RedditDB) UpdateVote(idPost int, vote *models.Vote) (*models.Post, error) {
	ctx := context.Background()

	var postExists bool
	err := s.db.QueryOne(ctx, &postExists, "SELECT EXISTS(SELECT 1 FROM Posts WHERE id = $1)", idPost)
	if err != nil {
		return nil, fmt.Errorf("ошибка при проверке существования поста: %w", err)
	}
	if !postExists {
		return nil, fmt.Errorf("пост с ID %d не найден", idPost)
	}

	err = s.db.WithTx(ctx, func(tx pg.Tx) error {
		var existingVote int
		var voteChange int

		queryCheck := `SELECT COALESCE(vote, 0) FROM Votes WHERE user_id = $1 AND post_id = $2`
		err := tx.QueryOne(ctx, &existingVote, queryCheck, vote.User, idPost)

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("ошибка при поиске голоса: %w", err)
		}

		if existingVote == vote.Vote {
			return nil
		}

		upsertVoteQuery := `
            INSERT INTO Votes (user_id, post_id, vote)
            VALUES ($1, $2, $3)
            ON CONFLICT (user_id, post_id) DO UPDATE SET vote = EXCLUDED.vote`
		_, err = tx.Exec(ctx, upsertVoteQuery, vote.User, idPost, vote.Vote)
		if err != nil {
			return fmt.Errorf("ошибка при обновлении/вставке голоса: %w", err)
		}

		voteChange = vote.Vote - existingVote

		if voteChange != 0 {
			postUpdate := `UPDATE Posts SET score = score + $1 WHERE id = $2`
			_, err = tx.Exec(ctx, postUpdate, voteChange, idPost)
			if err != nil {
				return fmt.Errorf("ошибка при обновлении score поста: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	updatedPost, err := s.GetPost(idPost)
	if err != nil {
		return nil, err
	}

	return updatedPost, nil
}
