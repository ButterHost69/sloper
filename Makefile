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
	go build -ldflags "$(go run ./tools/go-build-flags)" -o setup/sloper-web ./app/web
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

# ─── Dashboard ────────────────────────────────────────────────────────

# Build the per-instance read-only API server (app/web).
build-web:
	go build -ldflags "$(go run ./tools/go-build-flags)" -o dist/sloper-web ./app/web

# Install dashboard dependencies if they are missing or stale.
dashboard-install:
	cd dashboard/frontend && (test -x node_modules/.bin/next || (rm -rf node_modules && npm install --no-audit --no-fund))

# Install + build the Next.js console.
build-dashboard: dashboard-install
	cd dashboard/frontend && npm run build

# Serve the built console on port 3000 (default). Builds first if needed.
run-dashboard: dashboard-install
	cd dashboard/frontend && test -f .next/BUILD_ID || npm run build
	cd dashboard/frontend && npm start

# Next.js dev server with hot reload.
dev-dashboard: dashboard-install
	cd dashboard/frontend && npm run dev

dashboard: build-web build-dashboard