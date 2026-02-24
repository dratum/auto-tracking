package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Server    ServerConfig
	Timescale TimescaleConfig
	Mongo     MongoConfig
	Auth      AuthConfig
	Admin     AdminConfig
}

type ServerConfig struct {
	Host string `env:"SERVER_HOST" envDefault:"0.0.0.0"`
	Port int    `env:"SERVER_PORT" envDefault:"8080"`
}

type TimescaleConfig struct {
	Host     string `env:"TIMESCALE_HOST" envDefault:"localhost"`
	Port     int    `env:"TIMESCALE_PORT" envDefault:"5432"`
	User     string `env:"TIMESCALE_USER" envDefault:"autotrack"`
	Password string `env:"TIMESCALE_PASSWORD" envDefault:"secret"`
	DB       string `env:"TIMESCALE_DB" envDefault:"autotrack"`
	SSLMode  string `env:"TIMESCALE_SSLMODE" envDefault:"disable"`
}

type MongoConfig struct {
	URI string `env:"MONGO_URI" envDefault:"mongodb://localhost:27017"`
	DB  string `env:"MONGO_DB" envDefault:"autotrack"`
}

type AuthConfig struct {
	JWTSecret string        `env:"JWT_SECRET,required"`
	JWTExpiry time.Duration `env:"JWT_EXPIRY" envDefault:"168h"`
	APIKey    string        `env:"API_KEY,required"`
}

type AdminConfig struct {
	Username string `env:"DEFAULT_USERNAME" envDefault:"admin"`
	Password string `env:"DEFAULT_PASSWORD" envDefault:"admin"`
}

func (c TimescaleConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DB, c.SSLMode,
	)
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	return cfg, nil
}
