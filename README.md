# Forum Ahmed (Complete, No-JS)

A minimalist forum implemented in **Go**, **Go `html/template`**, pure **HTML/CSS**, and **SQLite**.  
**No JavaScript** and **no external CSS frameworks**.

## Features
- Register / Login / Logout (bcrypt, one active session per user, expiry)
- Create posts with multiple categories
- Browse posts with filters: by category, mine, liked
- Comments under posts
- Like/Dislike on posts and comments (mutually exclusive / toggles)
- Custom error pages (400, 401, 402, 404, 429, 500)
- Responsive, frosted-glass UI with gradient background and floating shapes
- Optional background images: add files to `static/images/` and reference in CSS if desired

## Run locally
```bash
make run
# or
go run main.go
```

## Docker
```bash
make build-docker
make run-docker
# open http://localhost:8080
```

## Project Structure
```
forum/
 ├── main.go
 ├── handlers/
 │    ├── auth.go
 │    ├── posts.go
 │    ├── comments.go (merged into posts.go for create)
 │    ├── likes.go
 │    ├── errors.go
 ├── models/
 │    ├── db.go
 │    ├── user.go
 │    ├── session.go
 │    ├── post.go
 │    ├── comment.go
 │    ├── like.go
 ├── templates/
 │    ├── layout.html
 │    ├── login.html
 │    ├── register.html
 │    ├── home.html
 │    ├── post.html
 │    ├── error.html
 ├── static/
 │    ├── style.css
 │    ├── images/
 ├── database/
 │    ├── schema.sql
 ├── Dockerfile
 ├── Makefile
 ├── README.md
 └── go.mod / go.sum
```

## Notes (ROSIC)
- **Languages**: only Go, Go templates, HTML, CSS, SQLite.
- **Allowed Go packages**: stdlib, `sqlite3`, `bcrypt`, `google/uuid`.
- **No JavaScript** anywhere; any timed UI fading uses CSS-only animations.
- **No frontend libraries** (React/Vue/etc).  
- **No external CSS frameworks**.  
- Everything renders via `html/template`.

## Audit checklist
- [x] Registration validates and prevents duplicates
- [x] Passwords hashed with bcrypt
- [x] Login starts single active session per user (expires)
- [x] Logout clears session
- [x] Only registered users can create posts/comments/react
- [x] Posts visible to all, with categories and filters
- [x] Comments shown with author/time
- [x] Likes/Dislikes on posts & comments, mutually exclusive
- [x] Custom error pages
- [x] Dockerfile builds, app runs on `:8080`
- [x] No JavaScript used


- Guest mode via `/guest` (view-only).
- Rate limit middleware returns custom 429 page.
