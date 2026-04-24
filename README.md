# 🚛 TrackOra — Enterprise GPS Fleet SaaS

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go)](https://go.dev)
[![Next.js](https://img.shields.io/badge/Next.js-14-black?logo=next.js)](https://nextjs.org)
[![Kafka](https://img.shields.io/badge/Apache_Kafka-3.x-231F20?logo=apachekafka)](https://kafka.apache.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?logo=postgresql)](https://postgresql.org)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

A **production-grade, multi-tenant GPS fleet tracking SaaS** built with 17 fully independent Go microservices, a Next.js 14 tenant dashboard, and a dedicated Super Admin panel.

---

## ✨ Platform Highlights

- **17 Go microservices** — each with its own PostgreSQL database, independently deployable
- **Zero-coupling architecture** — Apache Kafka for all cross-service communication; no direct HTTP
- **Real-time tracking** — MQTT ingest → WebSocket broadcast in < 100ms
- **Geofencing engine** — Ray-casting polygon evaluation on every location event
- **Rules-based alerts** — Speeding, geofence breach, maintenance overdue, offline device
- **Multi-tenant isolation** — Every tenant has dedicated database credentials and JWT scope
- **Clean Architecture** — Handler → Service → Repository → DB in every service
- **SQLC-ready SQL** — Raw `.sql` migrations per service, no ORM

---

## 🏗️ Architecture

```
┌───────────────────────────────────────────────────────────────────────────────┐
│                          CLIENTS                                              │
│   Tenant Dashboard (Next.js 14 :3000)    Admin Panel (Next.js 14 :3001)     │
└─────────────────────────┬────────────────────────────┬────────────────────────┘
                          │  REST / WebSocket           │  REST (Admin JWT)
                    ┌─────▼─────┐                       │
                    │  API      │                       │
                    │ Gateway   │◄──────────────────────┘
                    │  :8080    │
                    └─────┬─────┘
                          │  X-Tenant-Id / X-User-Id headers (JWT stripped here)
        ┌─────────────────┼─────────────────────────────────────────────────┐
        │                 │              Kafka Event Bus                    │
        │  ┌──────────────▼──────────────────────────────────────────────┐ │
        │  │  fleet.location.updated  fleet.trip.started  fleet.alert.*  │ │
        │  │  fleet.geofence.breach   fleet.trip.completed               │ │
        │  └──────────────┬──────────────────────────────────────────────┘ │
        │                 │                                                 │
  ┌─────▼─────┐   ┌──────▼──────┐   ┌─────────────┐   ┌─────────────────┐ │
  │ M01 Live  │   │ M03 Geo-   │   │ M04 Alerts  │   │ M02 Routes &   │ │
  │ Tracking  │   │ fencing    │   │             │   │ Trips           │ │
  │ :4001     │   │ :4003      │   │ :4004       │   │ :4002           │ │
  └─────┬─────┘   └──────┬──────┘   └──────┬──────┘   └───────┬─────────┘ │
        │ pg:5401         │ pg:5403          │ pg:5404           │ pg:5402  │
        └──═══════════════╧══════════════════╧═══════════════════╧══════════┘
                                     ...  and 13 more services
```

### Golden Rules
1. Each module = 1 Go microservice = 1 PostgreSQL database = 1 container
2. **No direct HTTP between services** — Apache Kafka for ALL cross-service events
3. API Gateway is the **ONLY** external entry point
4. JWT verified **only** at Gateway; services trust `X-User-Id` / `X-Tenant-Id` headers
5. Every service exposes `/health` and `/metrics` (Prometheus format)
6. Sony `gobreaker` circuit breaker on every downstream Gateway call
7. Clean Architecture: Handler → Service → Repository → DB in every service
8. SQLC for type-safe SQL — raw `.sql` files per service, no ORM

---

## 📦 Service Catalogue

| # | Service | Port | DB Port | Consumes | Produces |
|---|---------|------|---------|----------|----------|
| M01 | Live Tracking | 4001 | 5401 | (MQTT) | `location.updated` |
| M02 | Routes & Trips | 4002 | 5402 | `location.updated` | `trip.started`, `trip.completed` |
| M03 | Geofencing | 4003 | 5403 | `location.updated` | `geofence.breach` |
| M04 | Alerts | 4004 | 5404 | `location.updated`, `geofence.breach` | `alert.triggered` |
| M05 | Reports | 4005 | 5405 | — | — |
| M06 | Vehicles | 4006 | 5406 | — | `vehicle.created` |
| M07 | Drivers | 4007 | 5407 | — | `driver.created` |
| M08 | Maintenance | 4008 | 5408 | `trip.completed` | `maintenance.due` |
| M09 | Fuel | 4009 | 5409 | `trip.completed` | — |
| M10 | Multi-Tenant | 4010 | 5410 | — | `tenant.created` |
| M11 | Users & Access | 4011 | 5411 | — | `user.login` |
| M12 | Devices | 4012 | 5412 | — | `device.online`, `device.offline` |
| M13 | Security | 4013 | 5413 | `user.login`, `user.action` | — |
| M14 | Billing | 4014 | 5414 | — | `subscription.updated`, `invoice.created` |
| M15 | Admin Panel | 4015 | 5415 | — | — |
| M16 | Activity Log | 4016 | 5416 | `alert.triggered`, `trip.*` | — |
| M17 | Roadmap | 4017 | 5417 | — | — |

---

## 📁 Repository Structure

```
gpsgo/
├── go.work                          # Go workspace (17 modules)
├── docker-compose.yml               # Full infra stack
├── Makefile                         # Build & dev commands
│
├── shared/
│   ├── types/events.go              # All Kafka event structs
│   └── kafka/topics.go             # Topic name constants
│
├── protocols/
│   └── jt808/                      # JT808 vehicle protocol decoder
│
├── services/
│   ├── m01-live-tracking/          # MQTT + WebSocket + breadcrumbs
│   ├── m02-routes-trips/           # Trip FSM + route management
│   ├── m03-geofencing/             # Ray-cast polygon engine
│   ├── m04-alerts/                 # Rules evaluation engine
│   ├── m05-reports/                # Report defs + async PDF gen
│   ├── m06-vehicles/               # Vehicle CRUD
│   ├── m07-drivers/                # Driver CRUD + scoring
│   ├── m08-maintenance/            # Maintenance scheduler
│   ├── m09-fuel/                   # Fuel log management
│   ├── m10-multi-tenant/           # Tenant lifecycle
│   ├── m11-users-access/           # Auth + RBAC
│   ├── m12-devices/                # Device provisioning
│   ├── m13-security/               # Audit trail + incidents
│   ├── m14-billing/                # Stripe subscriptions
│   ├── m15-admin-panel/            # Super admin API (separate JWT)
│   ├── m16-activity-log/           # Activity feed (Kafka fan-in)
│   └── m17-roadmap/                # Feature voting board
│
├── frontend/
│   ├── tenant-dashboard/           # Next.js 14 (port 3000)
│   └── admin-panel/                # Next.js 14 (port 3001)
│
└── infra/
    ├── prometheus/prometheus.yml
    ├── grafana/
    ├── promtail/config.yml
    └── mosquitto/mosquitto.conf
```

---

## 🚀 Quick Start

### Prerequisites
- Docker Desktop ≥ 24
- Go 1.22+
- Node.js 20+

### 1. Start all infrastructure
```bash
docker compose up -d
```
This starts: Kafka (+ Zookeeper), 17 PostgreSQL instances, Mosquitto MQTT, MinIO, Prometheus, Grafana, Loki, Promtail.

### 2. Run database migrations

Each service has its own migrations in `services/<name>/db/migrations/`.  
Use [golang-migrate](https://github.com/golang-migrate/migrate) or run them manually:
```bash
# example for m01
psql "postgres://fleet:fleetpass@localhost:5401/fleet_tracking_db" \
  -f services/m01-live-tracking/db/migrations/000001_init.up.sql
```

### 3. Build all Go services
```bash
# Windows (PowerShell)
Get-ChildItem services -Directory | ForEach-Object {
  Push-Location $_.FullName
  go build ./cmd/...
  Pop-Location
}

# Linux / macOS
for d in services/*/; do (cd "$d" && go build ./cmd/...); done
```

### 4. Start services
```bash
# Example — start all in background (dev)
Get-ChildItem services -Directory | ForEach-Object {
  Start-Process -FilePath go -ArgumentList "run ./cmd/..." -WorkingDirectory $_.FullName
}
```

Or use Docker (recommended):
```bash
docker compose -f docker-compose.services.yml up -d
```

### 5. Start frontends
```bash
# Tenant Dashboard
cd frontend/tenant-dashboard && npm install && npm run dev   # http://localhost:3000

# Super Admin Panel
cd frontend/admin-panel    && npm install && npm run dev     # http://localhost:3001
```

---

## 🔑 Environment Variables

Each service reads via `internal/config/config.go`. Common variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | per-service | HTTP listen port |
| `DATABASE_URL` | per-service | PostgreSQL connection string |
| `KAFKA_BROKERS` | `localhost:29092` | Kafka bootstrap servers |
| `JWT_SECRET` | — | Tenant JWT (Gateway only) |
| `ADMIN_JWT_SECRET` | — | Super admin JWT (M15 only) |
| `MQTT_BROKER` | `localhost:1883` | MQTT broker (M01 only) |
| `MINIO_ENDPOINT` | `localhost:9000` | Object storage (M05 only) |
| `STRIPE_SECRET_KEY` | — | Stripe API key (M14 only) |

---

## 📡 API Endpoints

All tenant APIs go through the API Gateway (`localhost:8080`).  
Headers required: `Authorization: Bearer <tenant-jwt>` (Gateway verifies and adds `X-Tenant-Id`/`X-User-Id`).

### Sample: Vehicles (M06)
```http
GET    /api/vehicles          → list vehicles
POST   /api/vehicles          → create vehicle
GET    /api/vehicles/:id      → get vehicle
PUT    /api/vehicles/:id      → update vehicle
DELETE /api/vehicles/:id      → deactivate vehicle
```

### Sample: Live Tracking (M01)
```http
GET  /health                         → health check
WS   ws://localhost:4001/ws          → real-time location stream
GET  /api/tracking/:vehicle_id/breadcrumbs → trip breadcrumbs
POST /api/tracking/location          → manual location push (testing)
```

### Sample: Alerts (M04)
```http
GET    /api/alerts/rules             → list alert rules
POST   /api/alerts/rules             → create rule
GET    /api/alerts                   → list triggered alerts
POST   /api/alerts/:id/acknowledge   → acknowledge alert
POST   /api/alerts/:id/resolve       → resolve alert
```

### Admin Panel (M15)
```http
POST   /api/admin/auth/login         → get admin JWT (separate secret)
GET    /api/admin/tickets            → list support tickets (admin JWT required)
POST   /api/admin/tickets            → create ticket
```

---

## 📊 Observability

| Tool | URL | Purpose |
|------|-----|---------|
| Prometheus | `localhost:9090` | Metrics collection |
| Grafana | `localhost:3030` | Dashboards |
| Loki + Promtail | via Grafana | Centralized logs |

Every service exposes:
- `GET /health` → `{"status":"ok","service":"m01-live-tracking"}`
- `GET /metrics` → Prometheus text format

---

## 🧪 Testing

```bash
# Run geofencing engine tests (ray-cast + haversine)
cd services/m03-geofencing && go test ./internal/geom/... -v

# Run all tests across all services
Get-ChildItem services -Directory | ForEach-Object {
  Push-Location $_.FullName
  go test ./... 2>&1
  Pop-Location
}
```

**Test coverage:**
- ✅ `geom.Polygon.Contains` — 6 cases (inside, outside, edge, triangle, empty)
- ✅ `geom.Distance` — same point, known pair (Mumbai CST→Dadar), symmetry
- 🔲 Integration tests (services require running Postgres + Kafka)

---

## 🔧 Makefile Commands

```bash
make build-all       # Build all 17 services
make test            # Run all unit tests
make docker-up       # docker compose up -d
make docker-down     # docker compose down
make lint            # golangci-lint on all services
```

---

## 🌐 Frontend Pages

### Tenant Dashboard (`frontend/tenant-dashboard`)

| Route | Page |
|-------|------|
| `/dashboard` | Fleet overview — KPI cards, trip activity chart, alert breakdown |
| `/live-tracking` | Real-time map with animated vehicle markers + WebSocket |
| `/vehicles` | Vehicle table with search, filter, status badges |
| `/drivers` | Driver list with scoring bars |
| `/routes` | Trip history with from→to, distance, fuel |
| `/geofencing` | Zone map + geofence management |
| `/alerts` | Alert feed with severity filter + resolve actions |
| `/maintenance` | Scheduled task list with overdue indicators |
| `/fuel` | Fuel consumption chart + log table |
| `/reports` | Report definitions + async run history |
| `/devices` | Device fleet with signal strength bars |
| `/users` | Team member management + RBAC |
| `/security` | Audit log with action-color codes |
| `/billing` | Subscription plan + invoice table |
| `/activity` | Timeline activity feed (Kafka-powered) |
| `/roadmap` | Feature voting board (sorted by votes) |
| `/settings` | Company config, notification settings |

### Super Admin Panel (`frontend/admin-panel`)

| Route | Page |
|-------|------|
| `/dashboard` | MRR + tenant growth charts, service health grid |
| `/tenants` | Full tenant table with suspend/activate |
| `/tickets` | Support ticket queue with priority ranking |
| `/services` | All 17 services with latency + `/health` links |
| + more | billing, activity, security, roadmap, admins, announcements |

---

## 🔐 Security Model

```
Internet
  │
  ▼
API Gateway ── JWT Verification (HS256/RS256)
  │               └─ Extracts: tenant_id, user_id, role
  │               └─ Injects: X-Tenant-Id, X-User-Id, X-User-Role
  ▼
Microservices ── Trust Gateway headers; never verify JWT again
  │               └─ Tenant data isolation via tenant_id column filter
  │
M15 Admin Panel ── Separate JWT secret (ADMIN_JWT_SECRET)
  │               └─ Super admins cannot impersonate tenants
```

- Passwords hashed with **bcrypt** (cost 12)  
- All inter-service communication via **Kafka** (no shared secrets needed)
- Audit trails persisted by **M13 Security** (subscribes to `user.login` + `user.action`)

---

## 📋 Kafka Topics

| Topic | Producer | Consumers |
|-------|----------|-----------|
| `fleet.location.updated` | M01 | M02, M03, M04 |
| `fleet.trip.started` | M02 | M16 |
| `fleet.trip.completed` | M02 | M08, M09, M16 |
| `fleet.geofence.breach` | M03 | M04 |
| `fleet.alert.triggered` | M04 | M16 |
| `fleet.vehicle.created` | M06 | M12 |
| `fleet.driver.created` | M07 | — |
| `fleet.maintenance.due` | M08 | — |
| `fleet.device.online` | M12 | M01 |
| `fleet.tenant.created` | M10 | M11, M14 |
| `fleet.user.login` | M11 | M13 |
| `fleet.user.action` | M11 | M13 |
| `fleet.subscription.updated` | M14 | M10 |
| `fleet.invoice.created` | M14 | — |

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Follow Clean Architecture within the relevant service
4. Add unit tests for domain logic
5. Ensure `go build ./cmd/...` passes before opening a PR

---

## 📄 License

MIT © Trackora Technologies Pvt. Ltd.

