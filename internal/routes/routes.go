package routes

import (
	"fmt"
	"io/fs"
	"net/http"
	"reddit_v2"
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

	/*r := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./reddit_v2/static/"))
	r.Handle("/static/", http.StripPrefix("/static/", fileServer))*/

	/*r := http.NewServeMux()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/Users/sergejkrasnov/Desktop/reddit_v2/static/html/index.html")
	})
	r.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("/Users/sergejkrasnov/Desktop/reddit_v2/static"))))*/
	r := http.NewServeMux()

	// Обработка корневого пути
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// Отправляем содержимое index.html
		w.Write(reddit_v2.IndexHTML)

	})

	//Обработка статических файлов
	r.Handle("/static/", http.FileServer(http.FS(reddit_v2.StaticFiles)))
	err := fs.WalkDir(reddit_v2.StaticFiles, "static", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fmt.Println(path)
		return nil
	})

	if err != nil {
		fmt.Println("Ошибка:", err)
	}

	/*	r := http.NewServeMux()

		// Получаем текущую рабочую директорию
		currentDir, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		// Поднимаемся на один уровень вверх
		projectRoot := filepath.Join(currentDir, "..")

		// Строим путь к папке static
		staticPath := filepath.Join(projectRoot, "static")
		htmlPath := filepath.Join(projectRoot, "static", "html", "index.html")
		fmt.Println(staticPath)

		// Обработка корневого пути
		r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, htmlPath)
		})

		// Обработка статических файлов
		r.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))))
	*/
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
	authHandler.HandleFunc("/api/post/{"+PostID+"}/{"+CommentID+"}", userHandler.DeleteComment).Methods("DELETE")
	authHandler.HandleFunc("/api/post/{"+PostID+"}", userHandler.DeletePost).Methods("DELETE")
	authHandler.HandleFunc("/api/post/{"+PostID+"}/upvote", userHandler.Upvote).Methods("GET")
	authHandler.HandleFunc("/api/post/{"+PostID+"}/downvote", userHandler.Downvote).Methods("GET")
	authHandler.HandleFunc("/api/post/{"+PostID+"}/unvote", userHandler.Unvote).Methods("GET")
	return r
}

/*r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "/Users/sergejkrasnov/Desktop/reddit_v2/static/html/index.html")
})
r.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("/Users/sergejkrasnov/Desktop/reddit_v2/static"))))*/
