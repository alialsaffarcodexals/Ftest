
package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
)

// parse templates with shared layout
func parse(files ...string) (*template.Template, error) {
	var list []string
	list = append(list, filepath.Join("templates", "layout.html"))
	for _, f := range files {
		list = append(list, filepath.Join("templates", f))
	}
	return template.ParseFiles(list...)
}

// base data common to all pages
func baseData(r *http.Request) map[string]interface{} {
	data := map[string]interface{}{
		"UserID":    currentUserID(r),
		"Flash":     r.Header.Get("X-Flash"),
		"FlashType": r.Header.Get("X-Flash-Type"),
	}
	return data
}

// render helper that merges base data
func render(w http.ResponseWriter, r *http.Request, file string, data map[string]interface{}) {
	tpl, err := parse(file)
	if err != nil {
		RenderHTTPError(w, http.StatusInternalServerError, "Template error: "+err.Error())
		return
	}
	// merge
	base := baseData(r)
	for k, v := range data {
		base[k] = v
	}
	if err := tpl.ExecuteTemplate(w, "layout", base); err != nil {
		RenderHTTPError(w, http.StatusInternalServerError, "Template render error")
		return
	}
}
