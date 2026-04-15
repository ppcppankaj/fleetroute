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
RUN cd gateway && go build -ldflags="-s -w" -o gateway ./cmd/

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata wget
WORKDIR /app
COPY --from=builder /build/gateway .
EXPOSE 8000
HEALTHCHECK --interval=10s --timeout=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:8000/health || exit 1
ENTRYPOINT ["./gateway"]
