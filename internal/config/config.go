package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Environment string `yaml:"env" env:"ENV" env-default:"development"`
	LogLevel    string `yaml:"log_level" env:"LOG_LEVEL" env-default:"info"`

	GRPC struct {
		Host       string        `yaml:"host" env:"GRPC_HOST" env-default:"0.0.0.0"`
		Port       string        `yaml:"port" env:"GRPC_PORT" env-default:"50051"`
		Timeout    time.Duration `yaml:"timeout" env:"GRPC_TIMEOUT" env-default:"5s"`
		MaxConnAge time.Duration `yaml:"max_conn_age" env:"GRPC_MAX_CONN_AGE" env-default:"5m"`
	} `yaml:"grpc"`

	Gateway struct {
		Host           string        `yaml:"host" env:"GATEWAY_HOST" env-default:"0.0.0.0"`
		Port           string        `yaml:"port" env:"GATEWAY_PORT" env-default:"8080"`
		GRPCServerHost string        `yaml:"grpc_server_host" env:"GATEWAY_GRPC_SERVER_HOST" env-default:"localhost"`
		GRPCServerPort string        `yaml:"grpc_server_port" env:"GATEWAY_GRPC_SERVER_PORT" env-default:"50051"`
		Timeout        time.Duration `yaml:"timeout" env:"GATEWAY_TIMEOUT" env-default:"10s"`
	} `yaml:"gateway"`

	JWT struct {
		SecretKey     string        `yaml:"secret_key" env:"JWT_SECRET_KEY" env-default:"superpuper-secret-key"`
		TokenDuration time.Duration `yaml:"token_duration" env:"JWT_TOKEN_DURATION" env-default:"24h"`
	} `yaml:"jwt"`

	Swagger struct {
		AuthPath     string `yaml:"auth_path" env:"SWAGGER_AUTH_PATH" env-default:"./pkg/api/auth/auth.swagger.json"`
		ListingsPath string `yaml:"listings_path" env:"SWAGGER_LISTINGS_PATH" env-default:"./pkg/api/listings/listings.swagger.json"`
	} `yaml:"swagger"`

	Postgres struct {
		Host            string        `yaml:"host" env:"POSTGRES_HOST" env-default:"localhost"`
		Port            string        `yaml:"port" env:"POSTGRES_PORT" env-default:"5432"`
		User            string        `yaml:"user" env:"POSTGRES_USER" env-default:"postgres"`
		Password        string        `yaml:"password" env:"POSTGRES_PASSWORD" env-default:"postgres"`
		DB              string        `yaml:"db" env:"POSTGRES_DB" env-default:"marketplace"`
		SSLMode         string        `yaml:"ssl_mode" env:"POSTGRES_SSL_MODE" env-default:"disable"`
		MaxOpenConns    int           `yaml:"max_open_conns" env:"POSTGRES_MAX_CONNS" env-default:"25"`
		MinConns        int           `yaml:"min_conns" env:"POSTGRES_MIN_CONNS" env-default:"25"`
		ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env:"POSTGRES_CONN_MAX_LIFETIME" env-default:"5m"`
	} `yaml:"postgres"`

	Migrations struct {
		Dir string `yaml:"dir" env:"MIGRATIONS_DIR" env-default:"./migrations"`
	} `yaml:"migrations"`
}

func New() (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadConfig("./config/config.yml", cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// GetPostgresDSN возвращает строку подключения к PostgreSQL
func (c *Config) GetPostgresDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.Postgres.User,
		c.Postgres.Password,
		c.Postgres.Host,
		c.Postgres.Port,
		c.Postgres.DB,
		c.Postgres.SSLMode,
	)
}
