.PHONY: run build test migrate-up migrate-dn sqlc-gen swagger docker

run:
	go run ./cmd/main.go

build:
	go build -o bin/service ./cmd/main.go

test:
	go test ./... -v -cover

migrate-up:
	migrate -path db/migrations -database $$DATABASE_URL up

migrate-dn:
	migrate -path db/migrations -database $$DATABASE_URL down

sqlc-gen:
	sqlc generate

swagger:
	swag init -g cmd/main.go

docker:
	docker build -t m11-users-access .
