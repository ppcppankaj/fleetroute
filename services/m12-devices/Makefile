run:
	go run ./cmd/main.go

build:
	go build -o bin/m12-devices ./cmd/main.go

test:
	go test ./... -v -cover

migrate-up:
	migrate -path db/migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path db/migrations -database "$$DATABASE_URL" down
