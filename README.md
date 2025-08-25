# Forum Project

This repository contains the completed **forum_ahmed** project. The frontend templates and styles originate from the original `forum_ahmed` design, while the backend is integrated from the `forum-main` implementation.

## Building and Running

The project uses Docker for building and running the server. The provided Makefile offers convenient commands:

```sh
make build    # Build the Docker image
make run      # Run the container and expose the service on port 8080
```

Once running, the forum is available at [http://localhost:8080](http://localhost:8080).

## Project Structure

- `forum-ahmed/` – Source code of the forum
  - `cmd/server` – Application entry point
  - `internal/` – Backend packages and HTML templates
  - `data/` – SQLite database location
- `Dockerfile` – Multi-stage build configuration
- `Makefile` – Helper commands for building and running with Docker

The application uses SQLite for data storage and Go's standard library for serving HTTP requests.
