package main

import (
	"log"
	"net/http"
	"time"

	"forum/handlers"
	"forum/models"
)

func main() {
	// Init database + run schema
	if err := models.InitDB("forum.db", "database/schema.sql"); err != nil {
		log.Fatalf("db init: %v", err)
	}

	mux := http.NewServeMux()

	// Static assets
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Public pages
	mux.HandleFunc("/", handlers.Home) // list posts + filters
	mux.HandleFunc("/login", handlers.Login)
	mux.HandleFunc("/register", handlers.Register)
	mux.HandleFunc("/logout", handlers.Logout)
	mux.HandleFunc("/guest", handlers.Guest)

	// Posts & comments (auth required in handlers)
	mux.HandleFunc("/posts/create", handlers.CreatePost)
	mux.HandleFunc("/post/", handlers.ViewPost) // /post/{id}
	mux.HandleFunc("/comment/create", handlers.CreateComment)

	// Likes
	mux.HandleFunc("/reaction", handlers.Reaction) // like/dislike for posts & comments

	// Errors test endpoints (optional)
	mux.HandleFunc("/__400", func(w http.ResponseWriter, r *http.Request) { handlers.RenderHTTPError(w, 400, "Bad request (demo)") })
	mux.HandleFunc("/__401", func(w http.ResponseWriter, r *http.Request) { handlers.RenderHTTPError(w, 401, "Unauthorized (demo)") })
	mux.HandleFunc("/__402", func(w http.ResponseWriter, r *http.Request) { handlers.RenderHTTPError(w, 402, "Payment required (demo)") })
	mux.HandleFunc("/__404", func(w http.ResponseWriter, r *http.Request) { handlers.RenderHTTPError(w, 404, "Page not found (demo)") })
	mux.HandleFunc("/__429", func(w http.ResponseWriter, r *http.Request) { handlers.RenderHTTPError(w, 429, "Too many requests (demo)") })
	mux.HandleFunc("/__500", func(w http.ResponseWriter, r *http.Request) { handlers.RenderHTTPError(w, 500, "Internal server error (demo)") })

	srv := &http.Server{
	Addr: ":8080",
	Handler: handlers.WithTimeout( // <-- NEW
		handlers.WithRateLimit(
			handlers.WithSession(mux),
		),
	),
	ReadHeaderTimeout: 5 * time.Second,
}
	log.Println("Server starting on http://localhost:8080")
	log.Fatal(srv.ListenAndServe())
}

