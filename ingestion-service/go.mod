module gpsgo/ingestion-service

go 1.23.0

require (
	go.uber.org/zap v1.27.0
	gpsgo/pkg v0.0.0
	gpsgo/protocols v0.0.0
)

require (
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/nats-io/nats.go v1.35.0 // indirect
	github.com/nats-io/nkeys v0.4.7 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
)

replace (
	gpsgo/pkg => ../pkg
	gpsgo/protocols => ../protocols
)
