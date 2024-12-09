package storage

import (
	"fmt"
	"redditclone/internal/models"
	"redditclone/internal/token"
	"strconv"
	"sync"

	"github.com/golang-jwt/jwt"
)

type Interface interface {
	Register(user *models.User) (string, error)
	Login(user *models.User) (string, error)
	GetAllPosts() []*models.Post
	NewPost(post *models.Post) error
	GetPost(postID string) (*models.Post, error)
	GetPostsByCategory(category string) []*models.Post
	GetPostsByUserLogin(userLogin string) []*models.Post
	DeletePost(postID string) ([]*models.Post, error)
	UpdateVote(postID int, vote *models.Vote) *models.Post
	AddComment(postID int, comment *models.Comment) *models.Post
	DeleteComment(postID string, commentID string) *models.Post
	GetUser(userID int) string
}

type MemoryStorage struct {
	mu      sync.Mutex
	Users   map[string]*models.User
	Session map[string]string
	Posts   []*models.Post
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		Users:   make(map[string]*models.User),
		Session: make(map[string]string),
		Posts:   make([]*models.Post, 0),
	}
}

func (m *MemoryStorage) GetUser(userID int) string {
	for _, user := range m.Users {
		if user.ID == userID {
			return user.Username

		}
	}
	return "0"
}

func (m *MemoryStorage) Register(user *models.User) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, exists := m.Users[user.Username]
	if exists {
		return "", fmt.Errorf("пользователь уже существует")
	}

	user.ID = len(m.Users)

	m.Users[user.Username] = user

	jwt_token := jwt.NewWithClaims(jwt.SigningMethodHS256, token.GenerateTokenClaims(user))

	tokenString, err := jwt_token.SignedString(token.SecretKey)
	if err != nil {
		return "", fmt.Errorf("не удалось создать токен")
	}

	return tokenString, nil
}

func (m *MemoryStorage) Login(user *models.User) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	userLogin, exists := m.Users[user.Username]
	if !exists {
		return "", fmt.Errorf("пользователь не найден")
	}

	if userLogin.Password != user.Password {
		return "", fmt.Errorf("неверный пароль")
	}

	jwt_token := jwt.NewWithClaims(jwt.SigningMethodHS256, token.GenerateTokenClaims(user))

	tokenString, err := jwt_token.SignedString(token.SecretKey)
	if err != nil {
		return "", fmt.Errorf("не удалось создать токен")
	}

	m.Session[tokenString] = userLogin.Username // ??установка сессии

	return tokenString, nil

}
func (m *MemoryStorage) GetAllPosts() []*models.Post {
	return m.Posts
}

func (m *MemoryStorage) NewPost(post *models.Post) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	post.ID = len(m.Posts)
	m.Posts = append(m.Posts, post)

	return nil
}

func (m *MemoryStorage) GetPost(postID string) (*models.Post, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
		return nil, fmt.Errorf("не удалось распарсить пост")
	}

	for _, post := range m.Posts {
		if post.ID == postIDInt {
			return post, nil
		}
	}

	return nil, fmt.Errorf("пост не найден")
}

func (m *MemoryStorage) GetPostsByCategory(postCategory string) []*models.Post {
	m.mu.Lock()
	defer m.mu.Unlock()

	var postsCategory []*models.Post

	for _, post := range m.Posts {
		if post.Category == postCategory {
			postsCategory = append(postsCategory, post)
		}
	}
	return postsCategory
}

func (m *MemoryStorage) GetPostsByUserLogin(userLogin string) []*models.Post {
	m.mu.Lock()
	defer m.mu.Unlock()

	var postsUserLogin []*models.Post
	var userId int

	for _, user := range m.Users {
		if user.Username == userLogin {
			userId = user.ID
			break
		}
	}

	for _, post := range m.Posts {
		if post.Author.ID == userId {
			postsUserLogin = append(postsUserLogin, post)
		}
	}
	return postsUserLogin
}

func (m *MemoryStorage) DeletePost(postID string) ([]*models.Post, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить id поста")
	}
	for i, post := range m.Posts {
		if post.ID == postIDInt {
			m.Posts = append(m.Posts[:i], m.Posts[i+1:]...)
			return m.Posts, nil

		}
	}
	return nil, fmt.Errorf("не удалось удалить пост")
}

func (m *MemoryStorage) AddComment(postID int, comment *models.Comment) *models.Post {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, post := range m.Posts {
		if post.ID == postID {
			comment.ID = len(post.Comments)

			post.Comments = append(post.Comments, *comment)
			return post
		}
	}
	return nil
}

func (m *MemoryStorage) UpdateVote(postID int, vote *models.Vote) *models.Post {
	for _, post := range m.Posts {
		if post.ID == postID {
			for i, existingVote := range post.Votes {
				if existingVote.User == vote.User {
					post.Votes[i] = *vote

					k := 0
					for _, v := range post.Votes {
						k = k + v.Vote
					}
					post.Score = k
					return post
				}
			}
			post.Votes = append(post.Votes, *vote)
			k := 0
			for _, v := range post.Votes {
				k = k + v.Vote
			}
			post.Score = k
			return post
		}
	}
	return nil
}

func (m *MemoryStorage) DeleteComment(postID string, commentID string) *models.Post {
	m.mu.Lock()
	defer m.mu.Unlock()

	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
		return nil
	}
	commentIDInt, err := strconv.Atoi(commentID)
	if err != nil {
		return nil
	}
	for _, post := range m.Posts {
		if post.ID == postIDInt {
			for i, comment := range post.Comments {
				if comment.ID == commentIDInt {
					post.Comments = append(post.Comments[:i], post.Comments[i+1:]...)
					return post
				}
			}
		}
	}
	return nil
}
