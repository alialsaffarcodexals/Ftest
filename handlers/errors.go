// errors.go
package handlers

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
)

func RenderHTTPError(w http.ResponseWriter, code int, message string) {
	log.Printf("Rendering HTTP error: %d %s", code, message)
	// Try to render to a buffer first (so we don't send headers before we know it works)
	if tpl, err := parse("error.html"); err == nil {
		var buf bytes.Buffer
		if err2 := tpl.ExecuteTemplate(&buf, "layout", map[string]interface{}{
			"Title":   "Error",
			"Status":  code,
			"Message": message,
		}); err2 == nil {
			w.WriteHeader(code)
			_, _ = w.Write(buf.Bytes())
			return
		}
	}
	// Plain-text fallback that ALWAYS writes a body
	w.WriteHeader(code)
	_, _ = w.Write([]byte(fmt.Sprintf("%d %s\n%s", code, http.StatusText(code), message)))
}
