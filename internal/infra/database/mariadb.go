package database

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

const (
	mariaDBDriver           = "mysql" // MariaDB uses MySQL driver
	mariaDBConnectionString = "%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local"
)

func NewMariaDB(c *fiber.Ctx, cfg *SqlConfig) *Database {
	if len(cfg.Driver) == 0 {
		cfg.Driver = mariaDBDriver
	}

	return NewDatabase(c, cfg, mariaDBConnectionStringBuilder)
}

func mariaDBConnectionStringBuilder(cfg *SqlConfig) string {
	return fmt.Sprintf(mariaDBConnectionString, cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
}
