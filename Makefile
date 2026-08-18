.PHONY: build-sloper format-check lint build-all build-docker launch-docker docker docker-up docker-down docker-clean docker-clean-mem connect-docker

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


connect-docker:
	cd setup && sudo docker compose exec looper-service bash

docker: build-docker launch-docker

docker-up:
	cd setup && sudo docker compose up

docker-down:
	cd setup && sudo docker compose down

docker-clean:
	cd setup && sudo docker compose down -v
	sudo docker system prune -a --volumes -f

docker-clean-mem:
	sudo docker volume prune -f