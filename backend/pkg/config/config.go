package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	ServiceName string `mapstructure:"SERVICE_NAME"`
	GRPCPort    string `mapstructure:"GRPC_PORT"`
	HTTPPort    string `mapstructure:"HTTP_PORT"`

	DatabaseURL string `mapstructure:"DATABASE_URL"`
	RedisURL    string `mapstructure:"REDIS_URL"`

	KafkaBrokers []string `mapstructure:"KAFKA_BROKERS"`

	JWTSecret          string `mapstructure:"JWT_SECRET"`
	JWTAccessTokenTTL  int    `mapstructure:"JWT_ACCESS_TOKEN_TTL"`
	JWTRefreshTokenTTL int    `mapstructure:"JWT_REFRESH_TOKEN_TTL"`

	BookingMode string `mapstructure:"BOOKING_MODE"`

	UserServiceAddr    string `mapstructure:"USER_SERVICE_ADDR"`
	EventServiceAddr   string `mapstructure:"EVENT_SERVICE_ADDR"`
	BookingServiceAddr string `mapstructure:"BOOKING_SERVICE_ADDR"`
	PaymentServiceAddr string `mapstructure:"PAYMENT_SERVICE_ADDR"`
}

func Load() (*Config, error) {
	viper.AutomaticEnv()

	viper.SetDefault("GRPC_PORT", "50051")
	viper.SetDefault("HTTP_PORT", "8000")
	viper.SetDefault("REDIS_URL", "redis://localhost:6379")
	viper.SetDefault("KAFKA_BROKERS", []string{"localhost:9092"})
	viper.SetDefault("JWT_ACCESS_TOKEN_TTL", 15)
	viper.SetDefault("JWT_REFRESH_TOKEN_TTL", 7)
	viper.SetDefault("BOOKING_MODE", "pessimistic")

	cfg := &Config{}
	cfg.ServiceName = viper.GetString("SERVICE_NAME")
	cfg.GRPCPort = viper.GetString("GRPC_PORT")
	cfg.HTTPPort = viper.GetString("HTTP_PORT")
	cfg.DatabaseURL = viper.GetString("DATABASE_URL")
	cfg.RedisURL = viper.GetString("REDIS_URL")
	cfg.KafkaBrokers = viper.GetStringSlice("KAFKA_BROKERS")
	cfg.JWTSecret = viper.GetString("JWT_SECRET")
	cfg.JWTAccessTokenTTL = viper.GetInt("JWT_ACCESS_TOKEN_TTL")
	cfg.JWTRefreshTokenTTL = viper.GetInt("JWT_REFRESH_TOKEN_TTL")
	cfg.BookingMode = viper.GetString("BOOKING_MODE")
	cfg.UserServiceAddr = viper.GetString("USER_SERVICE_ADDR")
	cfg.EventServiceAddr = viper.GetString("EVENT_SERVICE_ADDR")
	cfg.BookingServiceAddr = viper.GetString("BOOKING_SERVICE_ADDR")
	cfg.PaymentServiceAddr = viper.GetString("PAYMENT_SERVICE_ADDR")

	return cfg, nil
}
