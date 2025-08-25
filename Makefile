IMAGE_NAME=forum-app

build:
	docker build -t $(IMAGE_NAME) .

run:
	docker run --rm -p 8080:8080 $(IMAGE_NAME)
