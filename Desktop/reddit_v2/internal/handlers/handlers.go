package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reddit_v2/internal/core"
	"reddit_v2/internal/middleware"
	"reddit_v2/internal/models"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
)

type UserHandler struct {
	service core.Interface
}

func NewUserHandler(service core.Interface) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var newUser models.User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	tokenString, err := h.service.Register(r.Context(), &newUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	cookie := &http.Cookie{
		Name:    "session_id",
		Value:   tokenString,
		Expires: time.Now().Add(12 * time.Hour),
	}
	http.SetCookie(w, cookie)

	response := middleware.RegisterResponse{AccessToken: tokenString}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var newUser models.User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	tokenString, err := h.service.Login(r.Context(), &newUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	cookie := &http.Cookie{
		Name:    "session_id",
		Value:   tokenString,
		Expires: time.Now().Add(12 * time.Hour),
	}
	http.SetCookie(w, cookie)

	response := middleware.RegisterResponse{AccessToken: tokenString}

	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) GetAllPosts(w http.ResponseWriter, r *http.Request) {
	//var allPosts []*models.Post
	allPosts, err := h.service.GetAllPosts(r.Context())
	/*for _, post := range h.service.GetAllPosts(r.Context()) {
		allPosts = append(allPosts, post)
	}*/
	if err != nil {
		http.Error(w, "Не удалось отправить посты", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(allPosts); err != nil {
		http.Error(w, "Не удалось сериализовать посты в JSON", http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie("session_id")
		if err != nil || cookie == nil {
			http.Error(w, "Токен не предоставлен", http.StatusUnauthorized)
			return
		}

		tokenString := cookie.Value

		claims := &middleware.TokenClaims{}
		jwt_token, err := jwt.ParseWithClaims(tokenString, claims, func(jwt_token *jwt.Token) (interface{}, error) {
			if _, ok := jwt_token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("неверный метод подписи")
			}
			return middleware.SecretKey, nil
		})

		if err != nil || !jwt_token.Valid {
			http.Error(w, "Неверный токен", http.StatusUnauthorized)
			return
		}

		if claims.EXP < time.Now().Unix() {
			http.Error(w, "Токен истек", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user_ID", claims.User.ID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (h *UserHandler) NewPost(w http.ResponseWriter, r *http.Request) {
	var newPost *models.Post
	err := json.NewDecoder(r.Body).Decode(&newPost)
	if err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	authorID, ok := r.Context().Value("user_ID").(int)
	if !ok {
		http.Error(w, "Не удалось получить ID автора", http.StatusUnauthorized)
		return
	}

	postAuthor := models.User{
		ID: authorID,
	}
	newPost.Author = postAuthor
	newPost.Created = time.Now()
	ctx := context.Background()

	if err := h.service.NewPost(ctx, newPost); err != nil {
		http.Error(w, "Не удалось создать пост", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newPost)
}

func (h *UserHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r) // мапа [POST_ID:1]
	idPost := vars["POST_ID"]

	post, err := h.service.GetPost(r.Context(), idPost)

	if err != nil {
		http.Error(w, "ошибка на стороне сервера", 400)
		return
	}

	json.NewEncoder(w).Encode(post)

}

func (h *UserHandler) GetPostsByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r) // мапа [CATEGORY_NAME:music] будет иметь ввид
	category := vars["CATEGORY_NAME"]
	posts, err := h.service.GetPostsByCategory(r.Context(), category)
	if err != nil {
		http.Error(w, "ошибка на стороне сервера", 400)
		return
	}
	json.NewEncoder(w).Encode(posts)

}

func (h *UserHandler) GetPostsByUserLogin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r) // мапа [CATEGORY_NAME:music] будет иметь ввид
	category := vars["USER_LOGIN"]
	posts, err := h.service.GetPostsByUserLogin(r.Context(), category)
	if err != nil {
		http.Error(w, "ошибка на стороне сервера", 400)
		return
	}
	json.NewEncoder(w).Encode(posts)
}

type CommentDTO struct {
	Body string `json:"comment"`
}

func (h *UserHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r) // мапа [POST_ID:1]
	idPost := vars["POST_ID"]

	var newComment models.Comment
	var newCommentDTO CommentDTO

	if err := json.NewDecoder(r.Body).Decode(&newCommentDTO); err != nil {
		http.Error(w, "Не удалось декодировать JSON", http.StatusBadRequest)
		return
	}

	authorID, ok := r.Context().Value("user_ID").(int)
	if !ok {
		http.Error(w, "Не удалось получить ID автора", http.StatusUnauthorized)
		return
	}

	userName, _ := h.service.GetUserName(r.Context(), authorID)

	newComment.Body = newCommentDTO.Body
	newComment.Author.ID = authorID
	newComment.Author.Username = userName
	newComment.Created = time.Now()

	post, err := h.service.AddComment(r.Context(), idPost, &newComment)
	if err != nil {
		http.Error(w, "Не удалось получить пост", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(post)
}

func (h *UserHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postIDStr := vars["POST_ID"]
	commentIDStr := vars["COMMENT_ID"]

	post, err := h.service.DeleteComment(r.Context(), postIDStr, commentIDStr)
	if err != nil {
		http.Error(w, "Не удалось получить пост", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(post)
}

func (h *UserHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID := vars["POST_ID"]

	postsAfterDeletion, err := h.service.DeletePost(r.Context(), postID)
	if err != nil {
		http.Error(w, "не удалось удалить пост", http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(postsAfterDeletion)
}

func (h *UserHandler) Upvote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r) // мапа [POST_ID:1]
	idPost := vars["POST_ID"]
	postID, err := strconv.Atoi(idPost)
	if err != nil {
		http.Error(w, "Неверный формат ID поста", http.StatusBadRequest)
		return
	}
	authorID, ok := r.Context().Value("user_ID").(int)
	if !ok {
		http.Error(w, "Не удалось получить ID автора", http.StatusUnauthorized)
		return
	}

	newVote := models.Vote{
		User: authorID,
		Vote: 1,
	}

	post, err := h.service.UpdateVote(r.Context(), postID, &newVote)
	if err != nil {
		http.Error(w, "не удалось отправить голос", http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(post)
}

func (h *UserHandler) Downvote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r) // мапа [POST_ID:1]
	idPost := vars["POST_ID"]
	postID, err := strconv.Atoi(idPost)
	if err != nil {
		http.Error(w, "Неверный формат ID поста", http.StatusBadRequest)
		return
	}
	authorID, ok := r.Context().Value("user_ID").(int)
	if !ok {
		http.Error(w, "Не удалось получить ID автора", http.StatusUnauthorized)
		return
	}

	newVote := models.Vote{
		User: authorID,
		Vote: -1,
	}
	post, err := h.service.UpdateVote(r.Context(), postID, &newVote)
	if err != nil {
		http.Error(w, "не удалось отправить голос", http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(post)
}

func (h *UserHandler) Unvote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r) // мапа [POST_ID:1]
	idPost := vars["POST_ID"]
	postID, err := strconv.Atoi(idPost)
	if err != nil {
		http.Error(w, "Неверный формат ID поста", http.StatusBadRequest)
		return
	}
	authorID, ok := r.Context().Value("user_ID").(int)
	if !ok {
		http.Error(w, "Не удалось получить ID автора", http.StatusUnauthorized)
		return
	}

	newVote := models.Vote{
		User: authorID,
		Vote: 0,
	}
	post, err := h.service.UpdateVote(r.Context(), postID, &newVote)
	if err != nil {
		http.Error(w, "не удалось отправить голос", http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(post)
}
