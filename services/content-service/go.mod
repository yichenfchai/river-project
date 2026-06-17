module github.com/grand-canal-guardian/services/content-service

go 1.22

require (
	github.com/grand-canal-guardian/pkg v0.0.0
	github.com/gin-gonic/gin v1.10.0
	github.com/google/uuid v1.6.0
	github.com/spf13/viper v1.19.0
	go.uber.org/zap v1.27.0
	gorm.io/driver/postgres v1.5.7
	gorm.io/gorm v1.25.10
)

replace github.com/grand-canal-guardian/pkg => ../../pkg
