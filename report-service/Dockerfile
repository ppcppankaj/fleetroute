FROM golang:1.22-alpine AS builder
WORKDIR /build
COPY go.work go.work.sum* ./
COPY pkg/ pkg/
COPY protocols/ protocols/
COPY api-service/ api-service/
COPY ingestion-service/ ingestion-service/
COPY stream-processor/ stream-processor/
COPY websocket-service/ websocket-service/
COPY maintenance-service/ maintenance-service/
COPY report-service/ report-service/
COPY gateway/ gateway/
COPY admin-panel/ admin-panel/
RUN cd report-service && go build -ldflags="-s -w" -o report-service ./cmd/

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /build/report-service .
# Local report output directory (used when S3 is not configured)
RUN mkdir -p /tmp/reports
ENTRYPOINT ["./report-service"]
