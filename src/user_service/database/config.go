package database

import (
	"github.com/jackc/pgx/v5"
)

type Config struct {
	DB      *pgx.Conn
	Queries *Queries
}

func GetConfig(conn *pgx.Conn) *Config {
	return &Config{
		DB:      conn,
		Queries: New(conn),
	}
}
