package database

import (
	"github.com/jackc/pgx/v5"
)

type Config struct {
	DB      *pgx.Conn
	Queries *Queries
}
