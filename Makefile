# ── GPS Fleet Management Platform — Makefile ─────────────────────────────────
# Works on Windows (PowerShell), Linux, and macOS.
# On Windows: run targets as   make <target>
# On Linux/macOS: same commands work natively.

SERVICES := ingestion-service stream-processor api-service websocket-service \
            maintenance-service report-service gateway admin-panel

# DB connection used for local migrate target (override via env or .env)
TIMESCALE_DSN ?= postgres://gpsgo:gpsgo@localhost:5432/gpsgo?sslmode=disable

# k6 test configuration
BASE_URL      ?= http://localhost:8000
AUTH_TOKEN    ?= replace-with-jwt
TENANT_ID     ?= replace-with-tenant-uuid
WS_URL        ?= ws://localhost:8000/ws

.PHONY: all build test lint tidy docker-up docker-down \
        migrate-up migrate-down migrate-create \
        frontend-install frontend-dev frontend-build \
        gen-keys clean help \
        k6-api k6-ws k6-ingestion k6-reports \
        helm-deploy helm-diff tf-plan tf-apply

all: build

## ── Help ──────────────────────────────────────────────────────────────────────
help:
	@echo "Available targets:"
	@echo "  docker-up        Start full dev stack (all services + infra)"
	@echo "  docker-infra     Start only DB + Redis + NATS"
	@echo "  docker-down      Stop and remove all containers + volumes"
	@echo "  migrate-up       Run all pending DB migrations (via Docker)"
	@echo "  migrate-down     Roll back the last migration"
	@echo "  frontend-install npm install in frontend/"
	@echo "  frontend-dev     Start Vite dev server on :5173"
	@echo "  gen-keys         Generate JWT RSA-4096 keypair into secrets/"
	@echo "  build            go build ALL 8 services into bin/"
	@echo "  test             go test -race all modules"
	@echo "  tidy             go mod tidy all modules"
	@echo "  k6-api           REST API load test (10k VUs)"
	@echo "  k6-ws            WebSocket load test (100k connections)"
	@echo "  k6-ingestion     TCP device simulator load test"
	@echo "  k6-reports       Report generation load test"
	@echo "  helm-deploy      Deploy to Kubernetes via Helm"
	@echo "  helm-diff        Diff Helm release (requires helm-diff plugin)"
	@echo "  tf-plan          terraform plan (infra/terraform/aws/)"
	@echo "  tf-apply         terraform apply (infra/terraform/aws/)"
	@echo "  clean            Remove bin/ and frontend/dist/"

## ── Docker ────────────────────────────────────────────────────────────────────

# Start the full stack — all Phase 3 services
docker-up:
	docker compose up -d

docker-build:
	docker compose build --parallel

docker-up-infra:
	docker compose up -d timescaledb redis nats prometheus grafana

# Wait for DB to be healthy then run migrations
migrate-up: docker-up
	@echo "Waiting for TimescaleDB to be ready..."
	docker compose run --rm migrate
	@echo "Migrations complete."

migrate-down:
	docker run --rm \
		-v "$(CURDIR)/migrations:/migrations" \
		migrate/migrate:v4.17.0 \
		-path=/migrations \
		-database="$(TIMESCALE_DSN)" \
		down 1

migrate-create:
	docker run --rm \
		-v "$(CURDIR)/migrations:/migrations" \
		migrate/migrate:v4.17.0 \
		create -ext sql -dir /migrations -seq $(NAME)

# Start infra only (skips app services that require Go build)
docker-infra:
	docker compose up -d timescaledb redis nats

docker-down:
	docker compose down -v

docker-logs:
	docker compose logs -f $(SERVICE)

## ── Go ────────────────────────────────────────────────────────────────────────

# On Windows without bash, run these in individual PowerShell sessions per module.
# On Linux/macOS these loop automatically.
build:
	@echo "Building all services..."
	cd pkg                && go build ./...
	cd protocols          && go build ./...
	cd ingestion-service  && go build -o ../bin/ingestion-service    ./cmd/...
	cd stream-processor   && go build -o ../bin/stream-processor     ./cmd/...
	cd api-service        && go build -o ../bin/api-service          ./cmd/...
	cd websocket-service  && go build -o ../bin/websocket-service    ./cmd/...
	cd maintenance-service && go build -o ../bin/maintenance-service ./cmd/...
	cd report-service     && go build -o ../bin/report-service       ./cmd/...
	cd gateway            && go build -o ../bin/gateway              ./cmd/...
	cd admin-panel        && go build -o ../bin/admin-panel          ./cmd/

test:
	cd pkg                && go test -race -count=1 ./...
	cd protocols          && go test -race -count=1 ./...
	cd ingestion-service  && go test -race -count=1 ./...
	cd stream-processor   && go test -race -count=1 ./...
	cd api-service        && go test -race -count=1 ./...
	cd maintenance-service && go test -race -count=1 ./...

test-coverage:
	cd api-service && go test -race -count=1 -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: api-service/coverage.html"

tidy:
	cd pkg                && go mod tidy
	cd protocols          && go mod tidy
	cd ingestion-service  && go mod tidy
	cd stream-processor   && go mod tidy
	cd api-service        && go mod tidy
	cd websocket-service  && go mod tidy
	cd maintenance-service && go mod tidy
	cd report-service     && go mod tidy
	cd gateway            && go mod tidy
	cd admin-panel        && go mod tidy

vet:
	cd pkg                && go vet ./...
	cd ingestion-service  && go vet ./...
	cd stream-processor   && go vet ./...
	cd api-service        && go vet ./...
	cd maintenance-service && go vet ./...
	cd report-service     && go vet ./...

## ── Frontend ──────────────────────────────────────────────────────────────────
frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

## ── Keys ──────────────────────────────────────────────────────────────────────
# Requires OpenSSL to be installed (comes with Git for Windows).
gen-keys:
	go build -o bin/gen_keys scripts/gen_keys.go
	bin/gen_keys
	@echo "JWT keys written to secrets/"

## ── Load Tests (requires k6 installed) ──────────────────────────────────────
k6-api:
	k6 run tests/k6/api_load.js \
		-e BASE_URL=$(BASE_URL) \
		-e AUTH_TOKEN=$(AUTH_TOKEN) \
		-e TENANT_ID=$(TENANT_ID)

k6-ws:
	k6 run tests/k6/websocket_load.js \
		-e WS_URL=$(WS_URL) \
		-e AUTH_TOKEN=$(AUTH_TOKEN)

k6-ingestion:
	k6 run --compatibility-mode=experimental_enhanced \
		tests/k6/ingestion_load.js \
		-e INGESTION_HOST=localhost

k6-reports:
	k6 run tests/k6/report_load.js \
		-e BASE_URL=$(BASE_URL) \
		-e AUTH_TOKEN=$(AUTH_TOKEN)

## ── Helm (requires kubectl + helm configured) ────────────────────────────────
helm-deploy:
	helm dependency update infra/k8s/helm/fleetos
	helm upgrade --install fleetos infra/k8s/helm/fleetos \
		--namespace fleetos --create-namespace \
		-f infra/k8s/helm/fleetos/values.yaml

helm-diff:
	helm diff upgrade fleetos infra/k8s/helm/fleetos \
		--namespace fleetos \
		-f infra/k8s/helm/fleetos/values.yaml

helm-uninstall:
	helm uninstall fleetos --namespace fleetos

## ── Terraform ────────────────────────────────────────────────────────────────
tf-plan:
	cd infra/terraform/aws && terraform init && terraform plan

tf-apply:
	cd infra/terraform/aws && terraform apply

tf-destroy:
	cd infra/terraform/aws && terraform destroy

## ── Cleanup ───────────────────────────────────────────────────────────────────
clean:
	@if exist bin  rmdir /s /q bin
	@if exist frontend\dist  rmdir /s /q frontend\dist
	@if exist admin-panel\uploads  rmdir /s /q admin-panel\uploads
