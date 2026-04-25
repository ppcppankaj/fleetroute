module gpsgo/gateway

go 1.23.0

require (
	github.com/go-chi/chi/v5 v5.1.0
	github.com/go-chi/cors v1.2.1
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/redis/go-redis/v9 v9.6.2
	github.com/segmentio/kafka-go v0.4.47
	github.com/sony/gobreaker v1.0.0
	gpsgo/shared v0.0.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	golang.org/x/text v0.23.0 // indirect
)

replace gpsgo/shared => ../shared
