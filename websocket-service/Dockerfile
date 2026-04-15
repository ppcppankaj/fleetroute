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
RUN cd websocket-service && go build -o /websocket-service ./cmd/...

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /websocket-service /websocket-service
EXPOSE 8081
ENTRYPOINT ["/websocket-service"]
