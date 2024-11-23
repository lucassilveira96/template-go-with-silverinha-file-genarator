package database

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

const (
	sqliteDriver           = "sqlite"
	sqliteConnectionString = "%s" // SQLite uses the file path as the DSN
)

func NewSQLite(c *fiber.Ctx, cfg *SqlConfig) *Database {
	if len(cfg.Driver) == 0 {
		cfg.Driver = sqliteDriver
	}

	return NewDatabase(c, cfg, sqliteConnectionStringBuilder)
}

func sqliteConnectionStringBuilder(cfg *SqlConfig) string {
	return fmt.Sprintf(sqliteConnectionString, cfg.Database) // Database is the file path for SQLite
}
