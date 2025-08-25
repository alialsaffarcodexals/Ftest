package handlers

import (
	"database/sql"
	"html/template"
)

type Handler struct {
	DB   *sql.DB
	Tmpl *template.Template
}

func NewHandler(db *sql.DB, tmpl *template.Template) *Handler {
	return &Handler{DB: db, Tmpl: tmpl}
}
