package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"forum-mvp/internal/app"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	addr := flag.String("addr", "8080", "http listen address")
	dataDir := "./data"
	tplDir := "./internal/web/templates"
	flag.Parse()

	// Ensure data dir exists
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		log.Fatal(err)
	}

	// Open/prepare SQLite DB
	dbPath := filepath.Join(dataDir, "forum.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	// Apply schema
	schema, err := os.ReadFile("internal/db/schema.sql")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		log.Fatal(err)
	}

	// Load HTML templates
	tpls, err := loadTemplates(tplDir)
	if err != nil {
		log.Fatal(err)
	}

	// App container
	a := &app.App{
		DB:         db,
		Templates:  tpls,
		CookieName: "forum_session",
		SessionTTL: 7 * 24 * time.Hour,
	}

	// Static assets (background, CSS, etc.)
	staticDir := filepath.Join(tplDir, "static")

	mux := http.NewServeMux()

	// Index route — only matches "/" now
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			// mark 404 → wrapper will render 404.html
			w.WriteHeader(http.StatusNotFound)
			return
		}
		a.HandleIndex(w, r)
	})

	// Routes
	mux.HandleFunc("/register", a.HandleRegister)
	mux.HandleFunc("/login", a.HandleLogin)
	mux.HandleFunc("/logout", a.HandleLogout)
	mux.HandleFunc("/post", a.HandleShowPost)
	mux.HandleFunc("/post/new", a.RequireAuth(a.HandleNewPost))
	mux.HandleFunc("/comment/new", a.RequireAuth(a.HandleNewComment))
	mux.HandleFunc("/like", a.RequireAuth(a.HandleLike))

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// Dev-only: trigger a panic to preview the 500 page
	mux.HandleFunc("/dev/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("intentional panic for 500 page preview")
	})

	// Wrap with custom error handler
	wrappedMux := withCustomErrors(mux, a)

	log.Printf("listening on :%s", *addr)
	log.Printf("Connect through: http://localhost:%s/", *addr)
	log.Fatal(http.ListenAndServe(":"+*addr, wrappedMux))
}

// Template loader
func loadTemplates(dir string) (map[string]*template.Template, error) {
	layout := filepath.Join(dir, "layout.html")
	pages, err := filepath.Glob(filepath.Join(dir, "*.html"))
	if err != nil {
		return nil, err
	}
	m := make(map[string]*template.Template)
	for _, page := range pages {
		if filepath.Base(page) == "layout.html" {
			continue
		}
		t, err := template.ParseFiles(layout, page)
		if err != nil {
			return nil, err
		}
		m[filepath.Base(page)] = t
	}
	return m, nil
}

// Custom error wrapper for 404/500 pages
func withCustomErrors(next *http.ServeMux, app *app.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 500 handler
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Header().Set("Cache-Control", "no-store")
				w.WriteHeader(http.StatusInternalServerError)
				if tpl, ok := app.Templates["500.html"]; ok {
					tpl.ExecuteTemplate(w, "500.html", appTemplateData(w, r, app))
				} else {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}
		}()

		// Capture status
		rw := &responseWriter{ResponseWriter: w, statusCode: 200}
		next.ServeHTTP(rw, r)

		// 404 handler
		if rw.statusCode == http.StatusNotFound {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "no-store")
			if tpl, ok := app.Templates["404.html"]; ok {
				tpl.ExecuteTemplate(w, "404.html", appTemplateData(w, r, app))
				return
			}
			http.NotFound(w, r)
		}
	})
}

// Capture status codes
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Base data passed to templates (user info, flash, etc.)
func appTemplateData(w http.ResponseWriter, r *http.Request, app *app.App) map[string]any {
	uid, uname, logged := app.CurrentUser(r)
	cats, _ := app.AllCategories()
	return map[string]any{
		"LoggedIn":   logged,
		"UserID":     uid,
		"Username":   uname,
		"Categories": cats,
		"Flash":      app.PopFlash(w, r),
	}
}
