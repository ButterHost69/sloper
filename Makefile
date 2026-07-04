.PHONY: build-sloper format-check lint build-all build-docker launch-docker docker docker-up docker-down

format-check:
	gofmt -l .

lint:
	go vet ./...

build-sloper:
	go build -ldflags "$(go run ./tools/go-build-flags)" -o dist/sloper ./app/sloper

build-all:
	go build -ldflags "$(go run ./tools/go-build-flags)" ./...

build-docker:
	go build -ldflags "$(go run ./tools/go-build-flags)" -o setup/sloper ./app/sloper
	cd setup && sudo docker compose build

launch-docker:
	cd setup && sudo docker compose up


docker: build-docker launch-docker

docker-up:
	cd sloper && sudo docker compose up

docker-down:
	cd setup && sudo docker compose down