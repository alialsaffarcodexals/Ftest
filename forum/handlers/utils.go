package handlers

import (
    "html/template"
    "net/http"
    "strconv"

    "forum/models"
)

// render parses templates and renders them with data.
func render(w http.ResponseWriter, tmpl string, data map[string]interface{}) {
    if data == nil {
        data = map[string]interface{}{}
    }
    t := template.Must(template.ParseFiles("templates/layout.html", "templates/"+tmpl))
    t.ExecuteTemplate(w, "layout", data)
}

// currentUser retrieves the logged in user from cookie.
func currentUser(r *http.Request) *models.User {
    c, err := r.Cookie("session")
    if err != nil {
        return nil
    }
    s, ok := models.GetSession(c.Value)
    if !ok {
        return nil
    }
    u, err := models.GetUserByID(s.UserID)
    if err != nil {
        return nil
    }
    return u
}

// parseID helper.
func parseID(s string) int {
    id, _ := strconv.Atoi(s)
    return id
}
