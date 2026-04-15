module gpsgo/maintenance-service

go 1.22

require (
	github.com/gin-gonic/gin v1.10.0
	github.com/jackc/pgx/v5 v5.6.0
	github.com/nats-io/nats.go v1.35.0
	github.com/redis/go-redis/v9 v9.5.3
	go.uber.org/zap v1.27.0
)

replace (
	gpsgo/pkg => ../pkg
)
