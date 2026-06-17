module github.com/grand-canal-guardian/services/api-gateway

go 1.22

require (
	github.com/grand-canal-guardian/pkg v0.0.0
	github.com/gin-gonic/gin v1.10.0
	github.com/redis/go-redis/v9 v9.5.3
	github.com/spf13/viper v1.19.0
	go.uber.org/zap v1.27.0
)

replace github.com/grand-canal-guardian/pkg => ../../pkg
