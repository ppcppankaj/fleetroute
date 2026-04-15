module gpsgo/report-service

go 1.22

require (
	github.com/aws/aws-sdk-go-v2 v1.30.0
	github.com/aws/aws-sdk-go-v2/config v1.27.0
	github.com/aws/aws-sdk-go-v2/service/s3 v1.57.0
	github.com/gin-gonic/gin v1.10.0
	github.com/jackc/pgx/v5 v5.6.0
	github.com/jung-kurt/gofpdf v1.16.2
	github.com/nats-io/nats.go v1.35.0
	go.uber.org/zap v1.27.0
)

replace (
	gpsgo/pkg => ../pkg
)
