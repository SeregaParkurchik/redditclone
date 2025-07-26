package middleware

import (
	"errors"
	"reddit_v2/internal/models"
	"time"
)

type TokenClaims struct {
	User struct {
		Username string `json:"username"`
		ID       int    `json:"id"`
	} `json:"user"`
	IAT int64 `json:"iat"`
	EXP int64 `json:"exp"`
}

type RegisterResponse struct {
	AccessToken string `json:"token"`
}

func GenerateTokenClaims(user *models.User) *TokenClaims {
	username := user.Username
	userID := user.ID

	newTokenClaims := &TokenClaims{
		User: struct {
			Username string `json:"username"`
			ID       int    `json:"id"`
		}{
			Username: username,
			ID:       userID,
		},
		IAT: time.Now().Unix(),
		EXP: time.Now().Add(time.Hour * 12).Unix(),
	}

	return newTokenClaims
}

func (c *TokenClaims) Valid() error {
	currentTime := time.Now().Unix()

	if c.EXP < currentTime {
		return errors.New("токен истек")
	}

	return nil
}

var SecretKey = []byte("мой ключ")
