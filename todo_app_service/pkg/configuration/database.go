package log

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

type databaseConfig struct {
	Port     string `envconfig:"DB_PORT"`
	User     string `envconfig:"DB_USER"`
	Password string `envconfig:"DB_PASS"`
	Host     string `envconfig:"DB_HOST"`
	Name     string `envconfig:"DB_NAME"`
	SslMode  string `envconfig:"DB_SSLMODE"`
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
