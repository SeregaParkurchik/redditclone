package core

import (
	"context"
	"redditclone/internal/models"
	"redditclone/internal/storage"
)

type Interface interface {
	Register(ctx context.Context, user *models.User) (string, error)
	Login(ctx context.Context, user *models.User) (string, error)
	GetAllPosts(ctx context.Context) []*models.Post
	NewPost(ctx context.Context, post *models.Post) error
	GetPost(ctx context.Context, postID string) (*models.Post, error)
	GetPostsByCategory(ctx context.Context, postCategory string) []*models.Post
	GetPostsByUserLogin(ctx context.Context, userLogin string) []*models.Post
	DeletePost(ctx context.Context, postID string) ([]*models.Post, error)
	UpdateVote(ctx context.Context, postID int, vote *models.Vote) *models.Post
	AddComment(ctx context.Context, postID int, comment *models.Comment) *models.Post
	//GetUser(ctx context.Context, userID int) string
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
	return s.storage.Register(user)
}

func (s *service) Login(ctx context.Context, user *models.User) (string, error) {
	return s.storage.Login(user)
}

func (s *service) GetAllPosts(ctx context.Context) []*models.Post {
	return s.storage.GetAllPosts()
}

func (s *service) NewPost(ctx context.Context, post *models.Post) error {
	return s.storage.NewPost(post)
}

func (s *service) GetPost(ctx context.Context, postID string) (*models.Post, error) {
	return s.storage.GetPost(postID)
}

func (s *service) GetPostsByCategory(ctx context.Context, postCategory string) []*models.Post {
	return s.storage.GetPostsByCategory(postCategory)
}

func (s *service) GetPostsByUserLogin(ctx context.Context, userLogin string) []*models.Post {
	return s.storage.GetPostsByUserLogin(userLogin)
}

func (s *service) DeletePost(ctx context.Context, postID string) ([]*models.Post, error) {
	return s.storage.DeletePost(postID)
}

func (s *service) AddComment(ctx context.Context, postID int, comment *models.Comment) *models.Post {
	return s.storage.AddComment(postID, comment)
}

func (s *service) UpdateVote(ctx context.Context, postID int, vote *models.Vote) *models.Post {
	return s.storage.UpdateVote(postID, vote)
}

/*func (s *service) GetUser(ctx context.Context, userID int) string {
	return s.storage.GetUser(userID)
}*/
