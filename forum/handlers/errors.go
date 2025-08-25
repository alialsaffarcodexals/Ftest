package handlers

import "net/http"

var errorMessages = map[int]string{
	400: "bad request",
	401: "unauthorized",
	402: "payment required",
	404: "page not found",
	429: "too many requests",
	500: "internal server error",
}

func (h *Handler) Error(w http.ResponseWriter, r *http.Request, code int) {
	msg := errorMessages[code]
	data := map[string]interface{}{
		"Title":   msg,
		"Theme":   h.theme(r),
		"Code":    code,
		"Message": msg,
	}
	w.WriteHeader(code)
	h.Tmpl.ExecuteTemplate(w, "error", data)
}
