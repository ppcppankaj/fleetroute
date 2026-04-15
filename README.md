# FleetOS — Enterprise GPS Fleet Management Platform

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go)](https://go.dev)
[![React](https://img.shields.io/badge/React-18-61DAFB?logo=react)](https://reactjs.org)
[![TimescaleDB](https://img.shields.io/badge/TimescaleDB-PostgreSQL-orange)](https://www.timescale.com)
[![NATS JetStream](https://img.shields.io/badge/NATS-JetStream-27AAE1)](https://nats.io)
[![AIS140](https://img.shields.io/badge/AIS140-VLTD%20Compliant-green)](https://morth.nic.in)

**FleetOS** is a production-grade, multi-tenant SaaS GPS tracking and fleet management platform built for commercial deployment. It supports **100,000+ concurrent devices**, white-labelling for resellers, and full compliance with India's **AIS140 / VLTD** standard.

---

## Table of Contents

- [Architecture](#architecture)
- [Features](#features)
- [Protocol Support](#protocol-support)
- [Services](#services)
- [Prerequisites](#prerequisites)
- [Quick Start (Docker Compose)](#quick-start-docker-compose)
- [Configuration](#configuration)
- [Database Migrations](#database-migrations)
- [Frontend Development](#frontend-development)
- [GoAdmin Panel](#goadmin-panel)
- [API Reference](#api-reference)
- [Load Testing](#load-testing)
- [Production Deployment](#production-deployment)
- [Observability](#observability)
- [Multi-Tenancy & Security](#multi-tenancy--security)
- [AIS140 Compliance](#ais140-compliance)
- [Project Structure](#project-structure)

---

## Architecture

```
                    ┌──────────────────────────────────────────────────────┐
                    │                     DEVICES                          │
                    │  Teltonika │ GT06 │ JT808 │ AIS140/VLTD │ TK103     │
                    └─────────┬────────────────────────────────────────────┘
                              │ TCP (per-protocol listeners)
                    ┌─────────▼──────────────────────────────────────────┐
                    │           INGESTION SERVICE  (:5008–5027)           │
                    │  • IMEI auth    • packet validation    • CRC check  │
                    │  • Dead-letter queue for malformed frames           │
                    └─────────┬──────────────────────────────────────────┘
                              │ NATS JetStream  (GPS_RAW)
                    ┌─────────▼──────────────────────────────────────────┐
                    │            STREAM PROCESSOR                         │
                    │  • Coordinate enrichment (PostGIS reverse-geocode)  │
                    │  • Trip FSM (start/end detection, distance / time)  │
                    │  • Geofence engine (polygon/circle/corridor PIP)    │
                    │  • Alert rule evaluator (JSONB condition trees)     │
                    │  • Driver score aggregation                         │
                    └────┬──────────────┬──────────────┬─────────────────┘
                         │GPS_ENRICHED  │ALERTS        │TRIPS
                ┌────────▼─────┐  ┌────▼────────┐  ┌──▼───────────┐
                │  TimescaleDB  │  │  NATS/Alert  │  │ Notification │
                │  avl_records  │  │  evaluator   │  │   Service    │
                │  (hypertable) │  └─────────────┘  └──────────────┘
                └──────────────┘
                         │
              ┌──────────▼─────────────────────────────────────────────┐
              │                    API SERVICE  (:8080)                  │
              │   REST: auth, devices, vehicles, drivers, geofences,     │
              │          alerts, trips, reports, maintenance, webhooks    │
              └──────────┬─────────────────────────────────────────────┘
                         │
              ┌──────────▼───────────┐     ┌────────────────────────┐
              │   WEBSOCKET SERVICE   │     │   REPORT SERVICE       │
              │   (:8081)             │     │   (NATS consumer)      │
              │   Fan-out per tenant  │     │   CSV / PDF → S3       │
              └──────────────────────┘     └────────────────────────┘
                         │
              ┌──────────▼───────────┐     ┌────────────────────────┐
              │      GATEWAY  (:8000) │     │   ADMIN PANEL  (:8090) │
              │   Reverse proxy + CORS│     │   GoAdmin + custom     │
              └──────────────────────┘     │   packet inspector      │
                                           └────────────────────────┘
```

### Data Stores

| Store         | Purpose                                      |
|---------------|----------------------------------------------|
| TimescaleDB   | `avl_records` hypertable, chunked by 1 week  |
| PostgreSQL RLS| All tenant tables with row-level security    |
| Redis         | Device state, WebSocket fan-out, rate limits |
| NATS JetStream| Event bus with durable consumers and DLQ     |
| S3            | Report file storage (CSV/PDF)                |

---

## Features

### Fleet Management
- ✅ Real-time live tracking — map with per-vehicle markers updated via WebSocket
- ✅ Route playback — timeline scrubber, speed-coloured path, 1×–60× speed
- ✅ Trip detection — automatic start/stop, distance, duration, overspeed events
- ✅ Geofence management — draw polygon, circle, corridor; entry/exit events; dwell analytics
- ✅ Driver management — RFID assignment, safety score (harsh accel/brake/corner)
- ✅ Vehicle management — registration, make/model/year, document tracking

### Maintenance
- ✅ Service schedules — time-based and odometer-based with proactive warnings
- ✅ Service log with technician, cost, and parts tracking
- ✅ Document management — FC, insurance, PUC, permits with expiry alerts
- ✅ Spare parts inventory with low-stock detection

### Alerting
- ✅ JSONB condition tree rule engine (AND/OR nesting, any telemetry field)
- ✅ Built-in templates — overspeed, geofence, idling, harsh driving, SOS, tamper
- ✅ Per-rule cooldown to prevent alert storms
- ✅ Multi-channel notifications — email, SMS, push, webhook

### Reporting
- ✅ 8 report types — trip, fuel, driver behaviour, geofence violations, idle, overspeed, maintenance, AIS140 audit
- ✅ Async generation — NATS queue → S3 upload → download link
- ✅ CSV and PDF output formats
- ✅ Scheduled reports with email delivery

### Platform
- ✅ Multi-tenant with full RLS isolation — zero cross-tenant data leakage
- ✅ JWT RS256 authentication with RBAC (17 fine-grained permissions)
- ✅ API key management for device/integration access
- ✅ Webhook delivery for third-party integration
- ✅ White-label branding per tenant (logo, primary colour, company name)
- ✅ GoAdmin operations panel — device inspector, protocol stats, NATS monitor, audit log

---

## Protocol Support

| Protocol | Port | Standard    | Packets/sec (per pod) |
|----------|------|-------------|----------------------|
| Teltonika Codec 8  | 5008 | Codec 8     | ~50,000 |
| Teltonika Codec 8E | 5008 | Codec 8 Ext | ~50,000 |
| GT06 / GT02        | 5023 | GT06        | ~40,000 |
| JT808 / JT809      | 5013 | JT808-2019  | ~40,000 |
| AIS140 / VLTD      | 5027 | AIS140:2016 | ~30,000 |
| TK103 / GPRMC      | 5018 | NMEA/ASCII  | ~60,000 |

---

## Services

| Service             | Port | Description                                      |
|---------------------|------|--------------------------------------------------|
| `ingestion-service` | 5008–5027 | TCP protocol handlers, IMEI auth, NATS publish  |
| `stream-processor`  | —    | NATS consumer; enrichment, trip FSM, geofence, alerts |
| `api-service`       | 8080 | REST API (Gin), JWT auth, 40+ endpoints          |
| `websocket-service` | 8081 | Real-time WebSocket push (tenant-isolated)       |
| `maintenance-service`| 8084 | Service schedules, documents, spare parts        |
| `report-service`    | 8085 | Async report generation (NATS worker), S3 upload |
| `gateway`           | 8000 | Reverse proxy + CORS                             |
| `admin-panel`       | 8090 | GoAdmin operations dashboard                     |
| `frontend`          | 5173 | React 18 + Vite SPA                              |

---

## Prerequisites

| Tool | Version |
|------|---------|
| Docker + Docker Compose | 24+ |
| Go | 1.22+ |
| Node.js | 20+ |
| OpenSSL (or LibreSSL) | 3.0+ |
| k6 (for load tests) | 0.51+ |
| Terraform | 1.6+ (for AWS) |
| Helm | 3.14+ (for K8s) |

---

## Quick Start (Docker Compose)

```bash
# 1. Clone the repository
git clone https://github.com/your-org/gpsgo.git
cd gpsgo

# 2. Generate RS256 JWT key pair
openssl genrsa -out keys/private.pem 4096
openssl rsa -in keys/private.pem -pubout -out keys/public.pem

# 3. Copy and configure environment
cp .env.example .env
# Edit .env — set TENANT_ID, etc.

# 4. Start all infrastructure + services
docker compose up -d

# 5. Run database migrations
docker compose exec api-service /app/api-service migrate

# 6. Start the frontend dev server
cd frontend
npm install
npm run dev
```

Access:
- **Frontend**: http://localhost:5173
- **API**: http://localhost:8000/api/v1
- **GoAdmin**: http://localhost:8090/admin  (user: `admin` / pass: `admin123`)
- **Grafana**: http://localhost:3000  (user: `admin` / pass: `admin`)
- **NATS Monitor**: http://localhost:8222

### Create First Tenant + Admin User

```bash
curl -X POST http://localhost:8000/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "admin@mycompany.com",
    "password": "SecurePass123!",
    "company_name": "My Fleet Company",
    "country": "IN"
  }'
```

---

## Configuration

All services read from environment variables. Copy `.env.example` to `.env`:

```bash
# Database
DATABASE_URL=postgres://gpsgo:gpsgo@localhost:5432/gpsgo?sslmode=disable

# NATS
NATS_URL=nats://localhost:4222

# Redis
REDIS_URL=redis://localhost:6379

# JWT keys
JWT_PRIVATE_KEY_PATH=./keys/private.pem
JWT_PUBLIC_KEY_PATH=./keys/public.pem

# Ingestion ports
TELTONIKA_PORT=5008
GT06_PORT=5023
JT808_PORT=5013
AIS140_PORT=5027
TK103_PORT=5018

# Report service
S3_BUCKET=gpsgo-reports
AWS_REGION=ap-south-1
REPORT_BASE_URL=http://localhost:8085/reports

# GoAdmin
ADMIN_PORT=8090
```

---

## Database Migrations

Migrations use sequential numbered SQL files. Run in order:

```bash
# With Docker
docker compose exec api-service /app/api-service migrate

# Manually (requires psql)
for f in migrations/*.up.sql; do
  psql "$DATABASE_URL" -f "$f"
done
```

| File | Contents |
|------|----------|
| `000001_schema.up.sql` | Core schema: tenants, users, devices |
| `000002_vehicles.up.sql` | Vehicles, drivers, trips, assignment |
| `000003_telemetry.up.sql` | `avl_records` TimescaleDB hypertable |
| `000004_geofences.up.sql` | Geofences, geofence_events, PostGIS |
| `000005_alerts.up.sql` | Alerts table |
| `000006_sessions.up.sql` | Auth sessions, refresh tokens |
| `000007_rules_notifications.up.sql` | ✨ Alert rules, notification channels, log |
| `000008_maintenance.up.sql` | ✨ Service schedules, docs, spare parts |
| `000009_report_jobs.up.sql` | ✨ Report queue, API keys, webhooks, packet_log |

---

## Frontend Development

```bash
cd frontend
npm install
npm run dev          # dev server on :5173
npm run build        # production build
npm run type-check   # TypeScript validation
```

### Tech stack
- **React 18** + TypeScript + Vite
- **TanStack Query** v5 — data fetching, cache
- **Zustand** — device state, alert store, auth
- **OpenLayers** — default map (swap to Mapbox/Google/MapmyIndia via adapter)
- **Recharts** — analytics charts
- **Lucide React** — icons

### Map Provider Swap

```typescript
// src/shared/map/MapAdapter.ts — switch provider in one line
import { GoogleMapsAdapter }    from './GoogleMapsAdapter'
import { MapboxAdapter  }       from './MapboxAdapter'
import { MapmyIndiaAdapter }    from './MapmyIndiaAdapter'
import { OpenLayersAdapter }    from './OpenLayersAdapter'  // default

export const MAP_PROVIDER = OpenLayersAdapter  // change here
```

---

## GoAdmin Panel

Access at **http://localhost:8090/admin** — internal operations only, not exposed publicly.

| Page | URL |
|------|-----|
| Dashboard (metrics + live chart) | `/admin/custom/dashboard` |
| Device Registry | `/admin/devices` |
| Tenant Management | `/admin/tenants` |
| Alert Rules | `/admin/alert_rules` |
| Alert History | `/admin/alerts_history` |
| Report Jobs | `/admin/report_jobs` |
| Raw Packet Inspector | `/admin/custom/packet-inspector` |
| Protocol Statistics | `/admin/custom/protocol-stats` |
| NATS Stream Monitor | `/admin/custom/nats-monitor` |
| Live Device Map | `/admin/custom/live-map` |
| Firmware Registry | `/admin/firmware_registry` |
| Admin Users | `/admin/admin_users` |
| Audit Log | `/admin/audit_log` |
| System Configuration | `/admin/system_config` |

---

## API Reference

All endpoints are prefixed `/api/v1`. Authentication: `Authorization: Bearer <jwt>`.

### Auth

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/auth/login` | Obtain JWT + refresh token |
| `POST` | `/auth/refresh` | Refresh JWT |
| `POST` | `/auth/register` | Create tenant + admin user |

### Devices

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/devices` | List all devices |
| `POST` | `/devices` | Register device |
| `GET` | `/devices/:id` | Device detail |
| `GET` | `/devices/:id/history` | Position history (time range) |
| `POST` | `/devices/:id/command` | Send remote command |

### Vehicles & Drivers

| Method | Path | Description |
|--------|------|-------------|
| `GET/POST` | `/vehicles` | List / create vehicles |
| `GET/PUT/DELETE` | `/vehicles/:id` | Vehicle CRUD |
| `GET/POST` | `/drivers` | List / create drivers |
| `GET` | `/drivers/:id/score` | Driver safety score |

### Geofences

| Method | Path | Description |
|--------|------|-------------|
| `GET/POST` | `/geofences` | List / create geofences |
| `DELETE` | `/geofences/:id` | Delete geofence |
| `GET` | `/geofences/:id/events` | Entry/exit events |

### Alerts

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/alerts` | List alerts (filterable by severity, type, date) |
| `POST` | `/alerts/:id/acknowledge` | Acknowledge alert |

### Reports

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/reports` | Submit async report job |
| `GET` | `/reports/:id` | Poll job status + download URL |
| `GET` | `/reports` | List report jobs |

### Maintenance

| Method | Path | Description |
|--------|------|-------------|
| `GET/POST` | `/maintenance/schedules` | Service schedules |
| `POST` | `/maintenance/schedules/:id/complete` | Mark service done |
| `GET/POST` | `/maintenance/documents` | Vehicle documents |
| `GET/POST` | `/maintenance/parts` | Spare parts inventory |

---

## Load Testing

```bash
# Install k6
choco install k6        # Windows
brew install k6         # macOS

# API load test (10,000 VUs)
k6 run tests/k6/api_load.js \
  -e BASE_URL=http://localhost:8000 \
  -e AUTH_TOKEN=<your-jwt>

# WebSocket test (100,000 concurrent connections)
k6 run tests/k6/websocket_load.js \
  -e WS_URL=ws://localhost:8081/ws

# Report generation under load
k6 run tests/k6/report_load.js

# TCP ingestion load (requires experimental net module)
k6 run --compatibility-mode=experimental_enhanced \
  tests/k6/ingestion_load.js \
  -e INGESTION_HOST=localhost
```

### Baseline performance targets

| Scenario | Target | Threshold |
|----------|--------|-----------|
| REST API p95 latency | < 200ms | < 500ms |
| REST API p99 latency | < 500ms | < 2000ms |
| WebSocket connections | 100,000 | error rate < 1% |
| Device ingestion throughput | 100,000 pkts/s | ACK rate > 99% |
| Report p95 completion | < 10s (CSV) | < 30s |

---

## Production Deployment

### AWS + EKS (recommended)

```bash
# 1. Provision infrastructure
cd infra/terraform/aws
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars — set db_password etc.
terraform init
terraform plan
terraform apply

# 2. Configure kubectl
aws eks update-kubeconfig --name gpsgo-prod --region ap-south-1

# 3. Create Kubernetes secrets
kubectl create secret generic gpsgo-db-secret \
  --from-literal=password='<DB_PASSWORD>'

kubectl create secret generic gpsgo-jwt-public-key \
  --from-file=public.pem=keys/public.pem

kubectl create secret generic gpsgo-jwt-private-key \
  --from-file=private.pem=keys/private.pem

# 4. Deploy with Helm
helm dependency update infra/k8s/helm/fleetos
helm upgrade --install fleetos infra/k8s/helm/fleetos \
  --namespace fleetos --create-namespace \
  --set global.imageRegistry=<ECR_URI> \
  --set externalDatabase.host=<RDS_ENDPOINT> \
  --set externalRedis.host=<ELASTICACHE_ENDPOINT> \
  -f infra/k8s/helm/fleetos/values.yaml
```

### Docker Compose (staging / single-server)

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

---

## Observability

| Tool | URL | Purpose |
|------|-----|---------|
| Grafana | :3000 | Dashboards — ingestion rates, API latency, DB connections |
| Prometheus | :9090 | Metrics scraping from all services |
| NATS Monitor | :8222 | JetStream stream stats |
| GoAdmin | :8090 | Operational insight — packets, errors, tenant usage |

### Key metrics exposed (`/metrics` on each service)

- `gpsgo_packets_received_total{protocol="teltonika"}` — ingestion rate
- `gpsgo_parse_errors_total{protocol}` — protocol error rate
- `gpsgo_device_connections_active` — active TCP connections
- `gpsgo_nats_publish_latency_seconds` — publish p50/p95/p99
- `gpsgo_db_query_duration_seconds` — query performance by endpoint
- `gpsgo_websocket_clients_active{tenant_id}` — WebSocket fan-out load
- `gpsgo_alerts_evaluated_total{rule_id}` — rule engine throughput

---

## Multi-Tenancy & Security

- **Row-Level Security (RLS)** — every query executes `SET LOCAL app.tenant_id = '<uuid>'`. The database enforces isolation; no application code can inadvertently leak cross-tenant data.
- **JWT RS256** — short-lived access tokens (15min), long-lived refresh tokens (30d). Private key never leaves the API service pod.
- **TLS 1.3** mandatory — enforced at the NLB/ALB layer and in docker-compose via TLS termination.
- **Audit log** — every mutation via GoAdmin is recorded with user, IP, entity type, and timestamp.
- **Rate limiting** — Redis sliding-window rate limiter per API key and per tenant.
- **Secret management** — Kubernetes secrets for DB password and JWT keys; never baked into images.

---

## AIS140 Compliance

FleetOS implements the full **AIS140:2016** (VLTD — Vehicle Location Tracking Device) standard:

| Requirement | Implementation |
|-------------|---------------|
| Mandatory fields | IMEI, timestamp, lat/lon, speed, heading, ignition, SOS, tamper, GPRS status |
| Emergency alert | SOS button → instant ALERTS.sos NATS event → notification within 3s |
| Immobilizer control | Remote command via `COMMANDS` NATS stream |
| Audit trail | AIS140 audit report type — full raw record export per vehicle |
| Driver ID | RFID-based driver assignment stored in `avl_records.driver_rfid` |
| Certificate | Firmware registry tracks AIS140-certified device models |
| Data retention | TimescaleDB continuous aggregate chunks — 2-year raw retention |

---

## Project Structure

```
gpsgo/
├── ingestion-service/          # TCP protocol listeners (5 protocols)
│   ├── cmd/                    # Main entry point
│   └── internal/
│       └── protocol/           # Teltonika, GT06, JT808, AIS140, TK103 handlers
│
├── stream-processor/           # NATS consumer: enrichment pipeline
│   └── internal/
│       ├── enrichment/         # Pipeline, geofence, trip FSM, alerts
│       └── state/              # Device FSM state machine
│
├── api-service/                # REST API (Gin)
│   └── internal/
│       ├── handler/            # 40+ HTTP handlers
│       ├── middleware/         # JWT auth, tenant context, rate limit
│       └── repository/         # pgx queries
│
├── websocket-service/          # Real-time push, NATS subscriber, Redis fan-out
│
├── maintenance-service/        # Service schedules, docs, spare parts API
│
├── report-service/             # Async report worker (NATS → CSV/PDF → S3)
│
├── gateway/                    # Reverse proxy + CORS
│
├── admin-panel/                # GoAdmin + custom pages (dashboard, inspector)
│   ├── cmd/                    # Main
│   ├── pages/                  # Custom GoAdmin pages
│   └── tables/                 # Table generators
│
├── pkg/                        # Shared Go packages
│   ├── logger/                 # zap wrapper
│   ├── metrics/                # Prometheus helpers
│   └── natsutil/               # JetStream helpers
│
├── frontend/                   # React 18 + TypeScript + Vite
│   └── src/
│       ├── features/
│       │   ├── tracking/       # LiveTracking (map)
│       │   ├── fleet/          # FleetPage, VehicleDetail, DriverDetail
│       │   ├── geofences/      # GeofencePage (draw + list + analytics)
│       │   ├── alerts/         # AlertsPage
│       │   ├── maintenance/    # MaintenancePage (4 tabs)
│       │   ├── reports/        # ReportsPage
│       │   ├── playback/       # RoutePlayback (timeline + telemetry)
│       │   ├── settings/       # SettingsPage (8 tabs)
│       │   └── auth/           # LoginPage
│       └── shared/
│           ├── map/            # OpenLayers adapter (swap to Google/Mapbox)
│           ├── store/          # Zustand: device, alert, auth
│           ├── websocket/      # useWebSocket hook
│           └── api/            # Axios client
│
├── migrations/                 # 9 numbered SQL migration files
│
├── infra/
│   ├── terraform/aws/          # VPC, EKS, RDS Aurora, ElastiCache, S3, NLB
│   └── k8s/helm/fleetos/       # Production Helm chart (HPA, PDB, NetworkPolicy)
│
├── tests/k6/                   # Load tests (API, WebSocket, ingestion, reports)
│
└── docker-compose.yml          # Full dev stack
```

---

## License

MIT — see [LICENSE](LICENSE)

---

## Commercial Support

For enterprise licensing, SLA support, custom protocol integration, or white-label deployment assistance, contact [hello@pankaj.im](mailto:hello@pankaj.im).
