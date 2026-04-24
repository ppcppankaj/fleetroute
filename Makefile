# ── GPS Fleet Management Platform — Makefile ─────────────────────────────────

SERVICES := m01-live-tracking m02-routes-trips m03-geofencing m04-alerts m05-reports m06-vehicles m07-drivers m08-maintenance m09-fuel m10-multi-tenant m11-users-access m12-devices m13-security m14-billing m15-admin-panel m16-activity-log m17-roadmap

.PHONY: all build test tidy docker-up docker-down migrate-up help clean

all: build

help:
	@echo "Available targets:"
	@echo "  docker-up        Start full dev stack (all services + infra)"
	@echo "  docker-down      Stop and remove all containers + volumes"
	@echo "  build            go build ALL 17 services into bin/"
	@echo "  test             go test -race all modules"
	@echo "  tidy             go mod tidy all modules"
	@echo "  frontend-install npm install in frontend/"
	@echo "  frontend-dev     Start Next.js dev server on :3001"
	@echo "  clean            Remove bin/ and frontend/.next/"

docker-up:
	docker compose up -d

docker-down:
	docker compose down -v

# Windows compatible build loops via powershell
build:
	@powershell -Command "if (!(Test-Path bin)) { New-Item -ItemType Directory -Path bin }; foreach (\$$s in '$(SERVICES)'.Split(' ')) { Write-Host \"Building \$$s...\"; cd services\\\$$s; go build -o ../../bin/\$$s.exe ./cmd/...; cd ../.. }; Write-Host \"Building gateway...\"; cd gateway; go build -o ../bin/gateway.exe ./cmd/...; cd .."

test:
	@powershell -Command "foreach (\$$s in '$(SERVICES)'.Split(' ')) { Write-Host \"Testing \$$s...\"; cd services\\\$$s; go test -race -count=1 ./...; cd ../.. }; Write-Host \"Testing gateway...\"; cd gateway; go test -race -count=1 ./...; cd .."

tidy:
	@powershell -Command "foreach (\$$s in '$(SERVICES)'.Split(' ')) { Write-Host \"Tidying \$$s...\"; cd services\\\$$s; go mod tidy; cd ../.. }; Write-Host \"Tidying gateway...\"; cd gateway; go mod tidy; cd .."

frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

admin-install:
	cd admin-panel && npm install

admin-dev:
	cd admin-panel && npm run dev

clean:
	@if exist bin rmdir /s /q bin
	@if exist frontend\.next rmdir /s /q frontend\.next
	@if exist admin-panel\.next rmdir /s /q admin-panel\.next
