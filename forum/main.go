package main

import (
    "log"
    "net/http"

    "github.com/gorilla/mux"

    "forum/handlers"
    "forum/models"
)

func main() {
    if err := models.InitDB("forum.db"); err != nil {
        log.Fatal(err)
    }
    defer models.Close()

    r := mux.NewRouter()
    r.HandleFunc("/", handlers.Index)
    r.HandleFunc("/register", handlers.Register)
    r.HandleFunc("/login", handlers.Login)
    r.HandleFunc("/logout", handlers.Logout)
    r.HandleFunc("/post/new", handlers.NewPost).Methods("GET", "POST")
    r.HandleFunc("/post/{id:[0-9]+}", handlers.ViewPost)
    r.HandleFunc("/post/{id:[0-9]+}/comment", handlers.AddComment).Methods("POST")
    r.HandleFunc("/post/{id:[0-9]+}/like", handlers.LikePostHandler).Methods("POST")
    r.HandleFunc("/comment/{id:[0-9]+}/like", handlers.LikeCommentHandler).Methods("POST")
    r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

    log.Println("Server running at :8080")
    if err := http.ListenAndServe(":8080", r); err != nil {
        log.Fatal(err)
    }
}
