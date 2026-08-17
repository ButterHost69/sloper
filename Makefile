.PHONY: build-sloper format-check lint build-all build-docker launch-docker docker docker-up docker-down docker-clean docker-clean-mem build-web build-dashboard run-dashboard dev-dashboard dashboard

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
	go build -ldflags "$(go run ./tools/go-build-flags)" -o setup/sloper-web ./app/web
	cd setup && sudo docker compose build

launch-docker:
	cd setup && sudo docker compose up


docker: build-docker launch-docker

docker-up:
	cd sloper && sudo docker compose up

docker-down:
	cd setup && sudo docker compose down

docker-clean:
	cd setup && sudo docker compose down -v
	sudo docker system prune -a --volumes -f

docker-clean-mem:
	sudo docker volume prune -f

# ─── Dashboard ────────────────────────────────────────────────────────

# Build the per-instance read-only API server (app/web).
build-web:
	go build -ldflags "$(go run ./tools/go-build-flags)" -o dist/sloper-web ./app/web

# Install + build the Next.js console.
build-dashboard:
	cd dashboard/frontend && npm install --no-audit --no-fund && npm run build

# Serve the built console on port 3000 (default).
run-dashboard:
	cd dashboard/frontend && npx next start -p 3000

# Next.js dev server with hot reload.
dev-dashboard:
	cd dashboard/frontend && npm run dev

dashboard: build-web build-dashboard