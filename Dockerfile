# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS builder
WORKDIR /app
RUN apk add --no-cache build-base
COPY forum-ahmed/forum/go.mod forum-ahmed/forum/go.sum ./
RUN go mod download
COPY forum-ahmed/forum .
RUN CGO_ENABLED=1 go build -o /out/forum ./cmd/server

FROM alpine:3.20
WORKDIR /srv
RUN mkdir -p /srv/data
COPY --from=builder /out/forum /srv/forum
COPY forum-ahmed/forum/internal /srv/internal
EXPOSE 8080
ENTRYPOINT ["/srv/forum","-addr=8080"]
