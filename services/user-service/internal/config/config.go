package config

import (
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port int    `mapstructure:"SERVER_PORT"`
	Mode string `mapstructure:"GIN_MODE"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"DB_HOST"`
	Port     int    `mapstructure:"DB_PORT"`
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASSWORD"`
	Name     string `mapstructure:"DB_NAME"`
}

func (c DatabaseConfig) DSN() string {
	// PostgreSQL DSN: host=... port=... user=... password=... dbname=... sslmode=disable
	return "host=" + c.Host + " port=" + itoa(c.Port) +
		" user=" + c.User + " password=" + c.Password +
		" dbname=" + c.Name + " sslmode=disable"
}

type RedisConfig struct {
	Host     string `mapstructure:"REDIS_HOST"`
	Port     int    `mapstructure:"REDIS_PORT"`
	Password string `mapstructure:"REDIS_PASSWORD"`
}

func (c RedisConfig) Addr() string {
	return c.Host + ":" + itoa(c.Port)
}

type JWTConfig struct {
	Secret          string `mapstructure:"JWT_SECRET"`
	AccessTokenTTL  int    `mapstructure:"JWT_ACCESS_TTL"`  // 秒, 默认 900
	RefreshTokenTTL int    `mapstructure:"JWT_REFRESH_TTL"` // 秒, 默认 604800
}

func Load() *Config {
	// 尝试从 .env 加载
	if _, err := os.Stat(".env"); err == nil {
		viper.SetConfigFile(".env")
		_ = viper.ReadInConfig()
	}
	viper.AutomaticEnv()

	// 设置默认值
	viper.SetDefault("SERVER_PORT", 8001)
	viper.SetDefault("GIN_MODE", "debug")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", 5432)
	viper.SetDefault("DB_USER", "canal")
	viper.SetDefault("DB_NAME", "grand_canal")
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", 6379)
	viper.SetDefault("JWT_ACCESS_TTL", 900)
	viper.SetDefault("JWT_REFRESH_TTL", 604800)

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		panic(err)
	}
	return cfg
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
