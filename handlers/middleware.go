
package handlers

import (
	"bytes"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"forum/models"
)

// ===== Flash (cookie-based) =====

const flashCookie = "flash" // value: kind|urlencoded(message)

func SetFlash(w http.ResponseWriter, kind, msg string) {
	if kind != "success" && kind != "error" {
		kind = "success"
	}
	v := kind + "|" + url.QueryEscape(msg)
	http.SetCookie(w, &http.Cookie{
		Name:     flashCookie,
		Value:    v,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   300, // 5min
		SameSite: http.SameSiteLaxMode,
	})
}

func getFlashFromCookie(w http.ResponseWriter, r *http.Request) (kind, msg string, ok bool) {
	c, err := r.Cookie(flashCookie)
	if err != nil || c.Value == "" {
		return "", "", false
	}
	parts := strings.SplitN(c.Value, "|", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	kind = parts[0]
	raw := parts[1]
	decoded, _ := url.QueryUnescape(raw)
	// clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     flashCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
	})
	return kind, decoded, true
}

// ===== Middlewares: RateLimit, Session, ErrorPages =====

// Simple fixed-window per-IP limiter
type rateWindow struct {
	start time.Time
	hits  int
}

var (
	rateMu      sync.Mutex
	rateBuckets = map[string]*rateWindow{}
	rateLimit   = 120 // requests
	rateWindowD = time.Minute
)

func clientIP(r *http.Request) string {
	// honor X-Forwarded-For if present (first)
	if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
		parts := strings.Split(xf, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func WithRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		now := time.Now()
		rateMu.Lock()
		b, ok := rateBuckets[ip]
		if !ok || now.Sub(b.start) > rateWindowD {
			b = &rateWindow{start: now, hits: 0}
			rateBuckets[ip] = b
		}
		b.hits++
		h := b.hits
		rateMu.Unlock()

		if h > rateLimit {
			log.Printf("Rate limit exceeded for %s", ip)
			RenderHTTPError(w, http.StatusTooManyRequests, "Too many requests. Please try again later.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// WithSession: refresh cookie + attach user id; also read any flash cookie into headers
func WithSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s", r.Method, r.URL.Path)
		// Flash: read & clear
		if kind, msg, ok := getFlashFromCookie(w, r); ok {
			r.Header.Set("X-Flash-Type", kind)
			r.Header.Set("X-Flash", msg)
		}
		// Session cookie
		if c, err := r.Cookie(sessionCookie); err == nil {
			if s, err := models.SessionFromID(c.Value); err == nil {
				http.SetCookie(w, &http.Cookie{
					Name:     sessionCookie,
					Value:    s.ID,
					Path:     "/",
					Expires:  time.Now().Add(models.SessionTTL),
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
				})
				r.Header.Set("X-User-ID", strconv64(s.UserID))
			}
		}
		next.ServeHTTP(w, r)
	})
}

// WithErrorPages captures 404 and replaces with our template.
type captureWriter struct {
	http.ResponseWriter
	status int
	buf    bytes.Buffer
}

func (cw *captureWriter) WriteHeader(code int) {
	cw.status = code
	// don't write yet
}

func (cw *captureWriter) Write(b []byte) (int, error) {
	return cw.buf.Write(b)
}


// Helper for int64 -> string without importing strconv in multiple files
func strconv64(v int64) string {
	return strconv.FormatInt(v, 10)
}


// middleware.go
// ...
func WithTimeout(next http.Handler) http.Handler {
	// 10s is generous for SQLite + templates; adjust if you prefer
	return http.TimeoutHandler(next, 10*time.Second, "Request timed out")
}
