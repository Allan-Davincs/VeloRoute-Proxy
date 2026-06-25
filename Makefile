.PHONY: dev test build lint run-backend run-frontend

dev:
	docker compose up --build

test:
	cd backend && go test ./... -v -race
	cd frontend && npm run lint

build:
	cd backend && go build -o bin/veloroute ./cmd/veloroute
	cd frontend && npm run build

lint:
	cd backend && go vet ./...
	cd frontend && npm run lint

run-backend:
	cd backend && go run ./cmd/veloroute --config config.yaml

run-frontend:
	cd frontend && npm run dev
