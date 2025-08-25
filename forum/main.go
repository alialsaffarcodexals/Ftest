package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	"forum/handlers"
	"forum/models"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := models.InitDB("database/forum.db", "database/schema.sql")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	tmpl := template.Must(template.ParseGlob("templates/*.html"))
	h := handlers.NewHandler(db, tmpl)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", h.Home)
	http.HandleFunc("/login", h.Login)
	http.HandleFunc("/register", h.Register)
	http.HandleFunc("/logout", h.Logout)
	http.HandleFunc("/post/", h.Post)
	http.HandleFunc("/comment", h.Comment)
	http.HandleFunc("/like", h.Like)
	http.HandleFunc("/toggle-theme", h.ToggleTheme)

	http.HandleFunc("/400", func(w http.ResponseWriter, r *http.Request) { h.Error(w, r, 400) })
	http.HandleFunc("/401", func(w http.ResponseWriter, r *http.Request) { h.Error(w, r, 401) })
	http.HandleFunc("/402", func(w http.ResponseWriter, r *http.Request) { h.Error(w, r, 402) })
	http.HandleFunc("/404", func(w http.ResponseWriter, r *http.Request) { h.Error(w, r, 404) })
	http.HandleFunc("/429", func(w http.ResponseWriter, r *http.Request) { h.Error(w, r, 429) })
	http.HandleFunc("/500", func(w http.ResponseWriter, r *http.Request) { h.Error(w, r, 500) })

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
