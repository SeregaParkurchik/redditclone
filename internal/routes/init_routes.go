package routes

import (
	"net/http"
	"redditclone/internal/handler"

	"github.com/gorilla/mux"
)

const (
	CategoryName = "CATEGORY_NAME"
	PostID       = "POST_ID"
	CommentID    = "COMMENT_ID"
	UserLogin    = "USER_LOGIN"
)

func InitRoutes(userHandler *handler.UserHandler) *http.ServeMux {
	r := http.NewServeMux()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/Users/sergejkrasnov/Desktop/redditclone/static/html/index.html")
	})

	r.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("/Users/sergejkrasnov/Desktop/redditclone/static"))))

	api := mux.NewRouter()
	r.Handle("/api/", api)

	api.HandleFunc("/api/register", userHandler.Register).Methods("POST")
	api.HandleFunc("/api/login", userHandler.Login).Methods("POST")
	api.HandleFunc("/api/posts/", userHandler.GetAllPosts).Methods("GET")
	api.HandleFunc("/api/posts/{"+CategoryName+"}", userHandler.GetPostsByCategory).Methods("GET")
	api.HandleFunc("/api/post/{"+PostID+"}", userHandler.GetPost).Methods("GET")
	api.HandleFunc("/api/user/{"+UserLogin+"}", userHandler.GetPostsByUserLogin).Methods("GET")

	authHandler := mux.NewRouter()
	authWithMiddlewareHandler := userHandler.AuthMiddleware(authHandler)
	api.PathPrefix("/api/").Handler(authWithMiddlewareHandler)

	authHandler.HandleFunc("/api/posts", userHandler.NewPost).Methods("POST")
	authHandler.HandleFunc("/api/post/{"+PostID+"}", userHandler.AddComment).Methods("POST")
	authHandler.HandleFunc("/api/post/{"+PostID+"}/{"+CommentID+"}", userHandler.DeleteComment).Methods("DELETE")
	authHandler.HandleFunc("/api/post/{"+PostID+"}/upvote", userHandler.Upvote).Methods("GET")
	authHandler.HandleFunc("/api/post/{"+PostID+"}/downvote", userHandler.Downvote).Methods("GET")
	authHandler.HandleFunc("/api/post/{"+PostID+"}/unvote", userHandler.Unvote).Methods("GET")

	authHandler.HandleFunc("/api/post/{"+PostID+"}", userHandler.DeletePost).Methods("DELETE")

	return r
}

//для mr
