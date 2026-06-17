package config

import (
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
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
	return "host=" + c.Host + " port=" + itoa(c.Port) +
		" user=" + c.User + " password=" + c.Password +
		" dbname=" + c.Name + " sslmode=disable"
}

func Load() *Config {
	if _, err := os.Stat(".env"); err == nil {
		viper.SetConfigFile(".env")
		_ = viper.ReadInConfig()
	}
	viper.AutomaticEnv()
	viper.SetDefault("SERVER_PORT", 8002)
	viper.SetDefault("GIN_MODE", "debug")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", 5432)
	viper.SetDefault("DB_USER", "canal")
	viper.SetDefault("DB_NAME", "grand_canal")

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
