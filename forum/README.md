# Forum

Simple forum web application written in Go using templates and SQLite.

## Setup

1. Ensure Go is installed.
2. Run database migrations and start the server:

```
make run
```

The server listens on `http://localhost:8080`.

## Features

- Registration and login with bcrypt password hashing.
- Create posts and comments.
- Like or dislike posts and comments.
- SQLite schema and migrations in `database/schema.sql`.
- Basic HTML templates and CSS without JavaScript.

## Docker

```
docker build -t forum .
docker run -p 8080:8080 forum
```
