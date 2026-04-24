FROM golang:1.23-alpine AS builder

WORKDIR /app

# Workspace definition
COPY go.work go.work.sum* ./

# Stub go.mod for every workspace module (needed for workspace resolution)
COPY pkg/go.mod pkg/go.sum* ./pkg/
COPY shared/go.mod shared/go.sum* ./shared/
COPY protocols/go.mod protocols/go.sum* ./protocols/
COPY gateway/go.mod gateway/go.sum* ./gateway/
COPY services/m01-live-tracking/go.mod services/m01-live-tracking/go.sum* ./services/m01-live-tracking/
COPY services/m02-routes-trips/go.mod services/m02-routes-trips/go.sum* ./services/m02-routes-trips/
COPY services/m03-geofencing/go.mod services/m03-geofencing/go.sum* ./services/m03-geofencing/
COPY services/m04-alerts/go.mod services/m04-alerts/go.sum* ./services/m04-alerts/
COPY services/m05-reports/go.mod services/m05-reports/go.sum* ./services/m05-reports/
COPY services/m06-vehicles/go.mod services/m06-vehicles/go.sum* ./services/m06-vehicles/
COPY services/m07-drivers/go.mod services/m07-drivers/go.sum* ./services/m07-drivers/
COPY services/m08-maintenance/go.mod services/m08-maintenance/go.sum* ./services/m08-maintenance/
COPY services/m10-multi-tenant/go.mod services/m10-multi-tenant/go.sum* ./services/m10-multi-tenant/
COPY services/m11-users-access/go.mod services/m11-users-access/go.sum* ./services/m11-users-access/
COPY services/m12-devices/go.mod services/m12-devices/go.sum* ./services/m12-devices/
COPY services/m13-security/go.mod services/m13-security/go.sum* ./services/m13-security/
COPY services/m14-billing/go.mod services/m14-billing/go.sum* ./services/m14-billing/
COPY services/m15-admin-panel/go.mod services/m15-admin-panel/go.sum* ./services/m15-admin-panel/
COPY services/m16-activity-log/go.mod services/m16-activity-log/go.sum* ./services/m16-activity-log/
COPY services/m17-roadmap/go.mod services/m17-roadmap/go.sum* ./services/m17-roadmap/

# Copy full source for THIS service
COPY shared/ ./shared/
COPY pkg/ ./pkg/
COPY protocols/ ./protocols/
COPY services/m09-fuel/ ./services/m09-fuel/

WORKDIR /app/services/m09-fuel
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/m09-fuel ./cmd/...

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/bin/m09-fuel ./service
EXPOSE 4001
CMD ["./service"]
