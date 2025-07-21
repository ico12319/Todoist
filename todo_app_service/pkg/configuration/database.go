package configuration

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

type databaseConfig struct {
	Port     string `envconfig:"POSTGRES_PORT" default:"5432"`
	User     string `envconfig:"POSTGRES_USER"`
	Password string `envconfig:"POSTGRES_PASSWORD"`
	Host     string `envconfig:"POSTGRES_HOST" default:"db"`
	Name     string `envconfig:"POSTGRES_DB"`
	SslMode  string `envconfig:"POSTGRES_SSLMODE" default:"disable"`
}

func OpenPostgres(cfg databaseConfig) *sqlx.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SslMode,
	)
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	return db
}
