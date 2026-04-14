module gpsgo/api-service

go 1.22

require (
	gpsgo/pkg v0.0.0
	github.com/gin-gonic/gin v1.10.0
	github.com/gin-contrib/cors v1.7.2
	github.com/jackc/pgx/v5 v5.6.0
	github.com/redis/go-redis/v9 v9.5.3
	github.com/swaggo/gin-swagger v1.6.0
	github.com/swaggo/swag v1.16.3
	go.uber.org/zap v1.27.0
	golang.org/x/crypto v0.23.0
)

replace gpsgo/pkg => ../pkg
