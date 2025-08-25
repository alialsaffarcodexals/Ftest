build:
	go build -o forum

run:
	go run main.go

build-docker:
	docker build -t forum .

run-docker:
	docker run --rm -p 8080:8080 forum
