
package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"forum/models"
)

func Reaction(w http.ResponseWriter, r *http.Request) {
	uid, ok := requireAuth(w, r); if !ok { return }
	if r.Method != http.MethodPost {
		RenderHTTPError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	typeStr := r.FormValue("target_type") // "post" or "comment"
	id, err := strconv.ParseInt(r.FormValue("target_id"), 10, 64)
	if err != nil {
		RenderHTTPError(w, http.StatusBadRequest, "Invalid target id")
		return
	}
	val, _ := strconv.Atoi(r.FormValue("value")) // 1 or -1
	if err := models.SetReaction(uid, strings.ToLower(typeStr), id, val); err != nil {
		SetFlash(w, "error", "Failed to react")
		// Try to redirect back to post
		if pid := r.FormValue("post_id"); pid != "" {
			http.Redirect(w, r, "/post/"+pid, http.StatusSeeOther); return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	SetFlash(w, "success", "Reaction saved")
	if typeStr == "post" {
		http.Redirect(w, r, "/post/"+strconv.FormatInt(id,10), http.StatusSeeOther)
		return
	}
	if pid := r.FormValue("post_id"); pid != "" {
		http.Redirect(w, r, "/post/"+pid, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
