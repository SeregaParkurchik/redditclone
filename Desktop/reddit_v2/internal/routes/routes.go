package routes

import (
	"net/http"
	"reddit_v2/internal/handlers"

	"github.com/gorilla/mux"
)

const (
	CategoryName = "CATEGORY_NAME"
	PostID       = "POST_ID"
	CommentID    = "COMMENT_ID"
	UserLogin    = "USER_LOGIN"
)

func InitRoutes(userHandler *handlers.UserHandler) *http.ServeMux {
	r := http.NewServeMux()

	/*fileServer := http.FileServer(http.Dir("./reddit_v2/static/"))
	r.Handle("/static/", http.StripPrefix("/static/", fileServer))*/
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/Users/sergejkrasnov/Desktop/reddit_v2/static/html/index.html")
	})
	r.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("/Users/sergejkrasnov/Desktop/reddit_v2/static"))))

	api := mux.NewRouter()
	r.Handle("/api/", api)

	api.HandleFunc("/api/register", userHandler.Register).Methods("POST")
	api.HandleFunc("/api/login", userHandler.Login).Methods("POST")
	api.HandleFunc("/api/posts/", userHandler.GetAllPosts).Methods("GET")
	api.HandleFunc("/api/post/{"+PostID+"}", userHandler.GetPost).Methods("GET")
	api.HandleFunc("/api/posts/{"+CategoryName+"}", userHandler.GetPostsByCategory).Methods("GET")
	api.HandleFunc("/api/user/{"+UserLogin+"}", userHandler.GetPostsByUserLogin).Methods("GET")

	authHandler := mux.NewRouter()
	authWithMiddlewareHandler := userHandler.AuthMiddleware(authHandler)
	api.PathPrefix("/api/").Handler(authWithMiddlewareHandler)

	authHandler.HandleFunc("/api/posts", userHandler.NewPost).Methods("POST")
	authHandler.HandleFunc("/api/post/{"+PostID+"}", userHandler.AddComment).Methods("POST")

	return r
}

/*r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "/Users/sergejkrasnov/Desktop/reddit_v2/static/html/index.html")
})
r.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("/Users/sergejkrasnov/Desktop/reddit_v2/static"))))*/
