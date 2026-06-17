package config

import (
	"os"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
}

type ServerConfig struct {
	Port int    `mapstructure:"SERVER_PORT"`
	Mode string `mapstructure:"GIN_MODE"`
}

type DatabaseConfig struct {
	Host, User, Password, Name string
	Port                       int
}

func (c DatabaseConfig) DSN() string {
	return "host=" + c.Host + " port=" + itoa(c.Port) +
		" user=" + c.User + " password=" + c.Password +
		" dbname=" + c.Name + " sslmode=disable"
}

type RedisConfig struct {
	Host     string `mapstructure:"REDIS_HOST"`
	Port     int    `mapstructure:"REDIS_PORT"`
	Password string `mapstructure:"REDIS_PASSWORD"`
}

func (c RedisConfig) Addr() string { return c.Host + ":" + itoa(c.Port) }

func Load() *Config {
	if _, err := os.Stat(".env"); err == nil {
		viper.SetConfigFile(".env"); _ = viper.ReadInConfig()
	}
	viper.AutomaticEnv()
	viper.SetDefault("SERVER_PORT", 8004)
	viper.SetDefault("GIN_MODE", "debug")
	viper.SetDefault("DB_HOST", "localhost"); viper.SetDefault("DB_PORT", 5432)
	viper.SetDefault("DB_USER", "canal"); viper.SetDefault("DB_NAME", "grand_canal")
	viper.SetDefault("REDIS_HOST", "localhost"); viper.SetDefault("REDIS_PORT", 6379)
	cfg := &Config{}
	_ = viper.Unmarshal(cfg)
	return cfg
}

func itoa(n int) string {
	if n == 0 { return "0" }
	s := ""
	for n > 0 { s = string(rune('0'+n%10)) + s; n /= 10 }
	return s
}
