package handlers

import "net/http"

// Error renders an error page with given status code.
func Error(w http.ResponseWriter, r *http.Request, code int, message string) {
    w.WriteHeader(code)
    data := map[string]interface{}{"Code": code, "Message": message, "User": currentUser(r)}
    render(w, "error.html", data)
}
