package core

import (
	"context"
	"fmt"
	"reddit_v2/internal/middleware"
	"reddit_v2/internal/models"
	"reddit_v2/internal/storage"
	"strconv"

	"github.com/golang-jwt/jwt"
)

type Interface interface {
	Register(ctx context.Context, user *models.User) (string, error)
	Login(ctx context.Context, user *models.User) (string, error)
	GetAllPosts(ctx context.Context) ([]*models.Post, error)
	NewPost(ctx context.Context, post *models.Post) error
	GetPost(ctx context.Context, post_ID string) (*models.Post, error)
	GetPostsByCategory(ctx context.Context, category string) ([]*models.Post, error)
	GetPostsByUserLogin(ctx context.Context, category string) ([]*models.Post, error)
	GetUserName(ctx context.Context, authorID int) (string, error)
	AddComment(ctx context.Context, idPost string, comment *models.Comment) (models.Post, error)
}

type service struct {
	storage storage.Interface
}

func New(storage storage.Interface) Interface {
	return &service{
		storage: storage,
	}
}

func (s *service) Register(ctx context.Context, user *models.User) (string, error) {
	// Хэшируем пароль перед сохранением
	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		return "", fmt.Errorf("не удалось хэшировать пароль: %w", err)
	}
	user.Password = hashedPassword // Сохраняем хэшированный пароль

	err = s.storage.Register(user)
	if err != nil {
		return "", err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, middleware.GenerateTokenClaims(user))
	tokenString, err := jwtToken.SignedString(middleware.SecretKey)
	if err != nil {
		return "", fmt.Errorf("не удалось создать токен")
	}

	return tokenString, nil
}

func (s *service) Login(ctx context.Context, user *models.User) (string, error) {
	foundUser, err := s.storage.Login(user)
	if err != nil {
		return "", err
	}

	if !CheckPasswordHash(user.Password, foundUser.Password) {
		return "", fmt.Errorf("неверный пароль")
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, middleware.GenerateTokenClaims(&foundUser))
	tokenString, err := jwtToken.SignedString(middleware.SecretKey)
	if err != nil {
		return "", fmt.Errorf("не удалось создать токен")
	}

	return tokenString, nil
}

func (s *service) GetAllPosts(ctx context.Context) ([]*models.Post, error) {
	posts, err := s.storage.GetAllPosts()

	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (s *service) NewPost(ctx context.Context, post *models.Post) error {
	err := s.storage.NewPost(post)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) GetPost(ctx context.Context, post_ID string) (*models.Post, error) {
	intPostID, err := strconv.Atoi(post_ID)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать преобразовать ID поста")
	}
	post, err := s.storage.GetPost(intPostID)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (s *service) GetPostsByCategory(ctx context.Context, category string) ([]*models.Post, error) {
	posts, err := s.storage.GetPostsByCategory(category)

	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (s *service) GetPostsByUserLogin(ctx context.Context, username string) ([]*models.Post, error) {
	posts, err := s.storage.GetPostsByUserLogin(username)

	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (s *service) GetUserName(ctx context.Context, authorID int) (string, error) {
	userName, err := s.storage.GetUserName(authorID)

	if err != nil {
		return "", err
	}
	return userName, nil
}

func (s *service) AddComment(ctx context.Context, idPost string, comment *models.Comment) (models.Post, error) {
	idPostINT, err := strconv.Atoi(idPost)

	if err != nil {
		return models.Post{}, err
	}
	post, err := s.storage.AddComment(idPostINT, comment)

	if err != nil {
		return models.Post{}, err
	}

	return post, nil
}
