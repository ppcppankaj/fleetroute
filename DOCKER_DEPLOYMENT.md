# TrackOra Platform — Docker Deployment Guide

**Company:** Trackora Technologies Pvt. Ltd.  
**Platform:** TrackOra — Enterprise GPS Fleet Tracking SaaS  
**Architecture:** 17 Go microservices + 2 Next.js 14 frontends on a single Docker network

---

## 🚀 One-Command Startup

```bash
docker compose up -d --build
```

That's it. The entire platform — 17 microservices, 17 PostgreSQL databases, Kafka, Redis, MQTT, MinIO, Prometheus, Grafana, and 2 Next.js frontends — starts with this single command.

---

## 📋 Prerequisites

| Tool | Minimum Version | Check |
|------|----------------|-------|
| **Docker Desktop** | 24.0+ | `docker --version` |
| **Docker Compose** | v2.20+ | `docker compose version` |
| **OpenSSL** | 3.x | `openssl version` |
| **RAM** | 8 GB free | Task Manager → Performance |
| **Disk** | 15 GB free | `df -h` |
| **OS** | Windows 10/11, macOS, Linux | — |

> **Windows users:** Enable WSL2 backend in Docker Desktop for best performance.

---

## 🏁 Step-by-Step First-Time Setup

### Step 1 — Clone / enter the project

```bash
cd C:\Users\ruchi\ppcp\gpsgo
```

### Step 2 — Generate JWT secrets (one-time only)

**Windows (PowerShell):**
```powershell
.\scripts\bootstrap-secrets.ps1
```

**Linux / macOS / WSL:**
```bash
chmod +x scripts/bootstrap-secrets.sh
./scripts/bootstrap-secrets.sh
```

This creates `secrets/jwt_private.pem` and `secrets/jwt_public.pem`.  
⚠️ **Never commit these files.** They are listed in `.gitignore`.

### Step 3 — Configure environment

```bash
cp .env.example .env
```

Edit `.env` and set at minimum:
- `ADMIN_JWT_SECRET` — a long random string (super admin login)
- `STRIPE_SECRET_KEY` — your Stripe key (or leave placeholder for dev)

### Step 4 — Start the full stack

```bash
docker compose up -d --build
```

**First run takes 5–10 minutes** while Docker:
- Pulls base images (Go 1.22, Node 20, Postgres 15, Kafka, etc.)
- Builds all 17 Go services from source
- Builds both Next.js apps (tenant dashboard + admin panel)
- Starts 50+ containers

---

## 🌐 Service URLs After Startup

| Service | URL | Credentials |
|---------|-----|-------------|
| **Tenant Dashboard** | http://localhost:3001 | — |
| **Super Admin Panel** | http://localhost:3002 | — |
| **API Gateway** | http://localhost:3000 | — |
| **Grafana** | http://localhost:3100 | admin / admin |
| **Prometheus** | http://localhost:9090 | — |
| **MinIO Console** | http://localhost:9091 | minioadmin / minioadmin123 |
| **Kafka** | localhost:29092 | — |
| **MQTT Broker** | localhost:1883 | — |

### Microservice Health Endpoints

| Service | Health URL |
|---------|-----------|
| M01 Live Tracking | http://localhost:4001/health |
| M02 Routes & Trips | http://localhost:4002/health |
| M03 Geofencing | http://localhost:4003/health |
| M04 Alerts | http://localhost:4004/health |
| M05 Reports | http://localhost:4005/health |
| M06 Vehicles | http://localhost:4006/health |
| M07 Drivers | http://localhost:4007/health |
| M08 Maintenance | http://localhost:4008/health |
| M09 Fuel | http://localhost:4009/health |
| M10 Multi-Tenant | http://localhost:4010/health |
| M11 Users & Access | http://localhost:4011/health |
| M12 Devices | http://localhost:4012/health |
| M13 Security | http://localhost:4013/health |
| M14 Billing | http://localhost:4014/health |
| M15 Admin Panel API | http://localhost:4015/health |
| M16 Activity Log | http://localhost:4016/health |
| M17 Roadmap | http://localhost:4017/health |

---

## 🛠️ Common Commands

```bash
# Start everything (first run or after changes)
docker compose up -d --build

# Start without rebuilding (fast restart)
docker compose up -d

# View all container status
docker compose ps

# Stream logs for all services
docker compose logs -f

# Stream logs for a specific service
docker compose logs -f m01-live-tracking
docker compose logs -f tenant-dashboard

# Restart a single service
docker compose restart m04-alerts

# Rebuild and restart one service only
docker compose up -d --build m06-vehicles

# Stop everything (keeps data volumes)
docker compose down

# Stop + wipe ALL data (fresh start)
docker compose down -v

# Check health of all services at once
docker compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"
```

---

## 📦 Container Map

### Infrastructure (always starts first)

| Container | Image | Purpose |
|-----------|-------|---------|
| `fleet-zookeeper` | confluentinc/cp-zookeeper:7.5.0 | Kafka coordination |
| `fleet-kafka` | confluentinc/cp-kafka:7.5.0 | Event bus |
| `fleet-redis` | redis:7-alpine | Rate limiting, caching |
| `fleet-mosquitto` | eclipse-mosquitto:2 | MQTT (vehicle GPS ingest) |
| `fleet-minio` | minio/minio | Object storage (reports, docs) |
| `fleet-prometheus` | prom/prometheus:v2.52.0 | Metrics collection |
| `fleet-grafana` | grafana/grafana:11.0.0 | Dashboards |
| `fleet-loki` | grafana/loki:2.9.0 | Log aggregation |
| `fleet-promtail` | grafana/promtail:2.9.0 | Log shipping |

### PostgreSQL Databases (one per service)

| Container | Port | Database |
|-----------|------|----------|
| `fleet-postgres-m01` | 5401 | fleet_tracking_db |
| `fleet-postgres-m02` | 5402 | fleet_routes_db |
| `fleet-postgres-m03` | 5403 | fleet_geofencing_db |
| `fleet-postgres-m04` | 5404 | fleet_alerts_db |
| `fleet-postgres-m05` | 5405 | fleet_reports_db |
| `fleet-postgres-m06` | 5406 | fleet_vehicles_db |
| `fleet-postgres-m07` | 5407 | fleet_drivers_db |
| `fleet-postgres-m08` | 5408 | fleet_maintenance_db |
| `fleet-postgres-m09` | 5409 | fleet_fuel_db |
| `fleet-postgres-m10` | 5410 | fleet_tenants_db |
| `fleet-postgres-m11` | 5411 | fleet_users_db |
| `fleet-postgres-m12` | 5412 | fleet_devices_db |
| `fleet-postgres-m13` | 5413 | fleet_security_db |
| `fleet-postgres-m14` | 5414 | fleet_billing_db |
| `fleet-postgres-m15` | 5415 | fleet_admin_db |
| `fleet-postgres-m16` | 5416 | fleet_activity_db |
| `fleet-postgres-m17` | 5417 | fleet_roadmap_db |

### Microservices

| Container | Port | Role |
|-----------|------|------|
| `fleet-gateway` | 3000 | API Gateway (only public entry point) |
| `fleet-m01` | 4001 | Live GPS tracking + WebSocket |
| `fleet-m02` | 4002 | Routes & trip FSM |
| `fleet-m03` | 4003 | Geofencing (ray-cast engine) |
| `fleet-m04` | 4004 | Rules-based alerts |
| `fleet-m05` | 4005 | Report generation |
| `fleet-m06` | 4006 | Vehicle management |
| `fleet-m07` | 4007 | Driver management |
| `fleet-m08` | 4008 | Maintenance scheduling |
| `fleet-m09` | 4009 | Fuel tracking |
| `fleet-m10` | 4010 | Multi-tenant management |
| `fleet-m11` | 4011 | Users & RBAC |
| `fleet-m12` | 4012 | Device provisioning |
| `fleet-m13` | 4013 | Security & audit |
| `fleet-m14` | 4014 | Billing (Stripe) |
| `fleet-m15` | 4015 | Super admin API |
| `fleet-m16` | 4016 | Activity log feed |
| `fleet-m17` | 4017 | Feature roadmap |

### Frontends

| Container | Exposed Port | App |
|-----------|-------------|-----|
| `trackora-tenant-dashboard` | 3001 | Tenant Fleet Dashboard |
| `trackora-admin-dashboard` | 3002 | Super Admin Panel |

---

## 🗂️ Data Volumes

All data is persisted in named Docker volumes. They survive `docker compose down` but are removed with `docker compose down -v`.

```bash
# List all TrackOra volumes
docker volume ls | grep pgdata
docker volume ls | grep fleet

# Inspect a specific volume
docker volume inspect gpsgo_pgdata_m06
```

**Volume names:** `pgdata_m01` → `pgdata_m17`, `redis_data`, `kafka_data`, `minio_data`, `grafana_data`, `prometheus_data`

---

## 🔐 Secrets Management

Secrets are mounted as Docker secrets from `./secrets/`:

```
secrets/
  jwt_private.pem   ← RSA private key (signs JWT tokens in Gateway)
  jwt_public.pem    ← RSA public key (verifies in Gateway)
```

- Services `m11-users-access` and `gateway` mount these as `/run/secrets/`
- M15 Admin Panel uses `ADMIN_JWT_SECRET` env var (separate from tenant JWTs)

> **Production:** Replace file-based secrets with Docker Swarm secrets or HashiCorp Vault.

---

## 🔍 Startup Sequence

Docker Compose respects `depends_on` health checks to boot in order:

```
zookeeper (healthy)
    └─ kafka (healthy)
           └─ All 17 microservices (healthy)
                  └─ gateway (healthy)
                         └─ tenant-dashboard, admin-dashboard

postgres-m01..17 (healthy)
    └─ respective microservices
```

> **Tip:** Run `docker compose ps` after 2–3 minutes to check all health statuses show `(healthy)`.

---

## 📊 Observability

### Prometheus Metrics
Every service exposes `GET /metrics` in Prometheus format.  
Prometheus scrapes all 17 services + gateway automatically via `infra/prometheus/prometheus.yml`.

### Grafana Dashboards
1. Open http://localhost:3100
2. Login: `admin` / `admin`
3. Dashboards are auto-provisioned from `infra/grafana/provisioning/`

### Logs
Promtail ships all container logs to Loki. View in Grafana:  
**Explore → Loki → `{container_name="fleet-m04"}`**

---

## 🐞 Troubleshooting

### "Port already in use"
```bash
# Find what's using port 3000
netstat -ano | findstr :3000       # Windows
lsof -i :3000                       # Linux/macOS

# Stop the conflicting process, then:
docker compose up -d
```

### Kafka keeps restarting
```bash
# Zookeeper must be healthy first
docker compose logs zookeeper
# Wait 60s then try again:
docker compose restart kafka
```

### Service runs but returns 500
```bash
# Check if its database ran migrations
docker compose logs m06-vehicles
# Look for "connection refused" or "relation does not exist"
```

### Out of memory
Reduce concurrency by starting infra first, then services:
```bash
docker compose up -d zookeeper kafka redis postgresql-m01..m17
sleep 30
docker compose up -d   # now start the rest
```

### Fresh restart (wipe everything)
```bash
docker compose down -v
docker compose up -d --build
```

---

## 🏭 Production Checklist

- [ ] Replace `ADMIN_JWT_SECRET` with a cryptographically random 64-char string
- [ ] Add real `STRIPE_SECRET_KEY`
- [ ] Configure `SMTP_*` for email alerts
- [ ] Change default Grafana admin password
- [ ] Change MinIO root credentials
- [ ] Set PostgreSQL passwords to strong values (update all `DATABASE_URL` envs)
- [ ] Mount TLS certificates for HTTPS
- [ ] Set `KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=3` for multi-broker setup
- [ ] Use external managed PostgreSQL (AWS RDS, Supabase) for production
- [ ] Enable `NEXT_TELEMETRY_DISABLED=1` ✅ (already set in Dockerfiles)
- [ ] Run secrets through Docker Swarm secrets or Vault instead of files

---

## 📁 Key Files Reference

```
gpsgo/
├── docker-compose.yml              ← Single source of truth for all containers
├── .env.example                    ← Copy to .env and fill in values
├── .env                            ← Your local config (gitignored)
├── secrets/                        ← JWT keys (gitignored)
│   ├── jwt_private.pem
│   └── jwt_public.pem
├── scripts/
│   ├── bootstrap-secrets.sh        ← Linux/macOS key generation
│   └── bootstrap-secrets.ps1       ← Windows key generation
├── infra/
│   ├── prometheus/prometheus.yml   ← Prometheus scrape config
│   ├── grafana/provisioning/       ← Grafana dashboards & datasources
│   ├── promtail/config.yml         ← Log shipping config
│   └── mosquitto/mosquitto.conf    ← MQTT broker config
├── frontend/
│   ├── tenant-dashboard/           ← Next.js 14 fleet dashboard (port 3001)
│   │   ├── Dockerfile
│   │   └── next.config.js          ← output: 'standalone'
│   └── admin-panel/                ← Next.js 14 super admin (port 3002)
│       ├── Dockerfile
│       └── next.config.js          ← output: 'standalone'
└── services/
    └── m01-live-tracking/
    └── ... (m02 through m17)
        └── Dockerfile              ← Multi-stage Go build
```

---

*TrackOra by Trackora Technologies Pvt. Ltd. — Built with ❤️ using Go, Next.js, Kafka, and PostgreSQL.*
