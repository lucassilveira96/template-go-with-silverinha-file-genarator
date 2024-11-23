package database

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

const (
	mysqlDriver           = "mysql"
	mysqlConnectionString = "%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local"
)

func NewMySQL(c *fiber.Ctx, cfg *SqlConfig) *Database {
	if len(cfg.Driver) == 0 {
		cfg.Driver = mysqlDriver
	}

	return NewDatabase(c, cfg, mysqlConnectionStringBuilder)
}

func mysqlConnectionStringBuilder(cfg *SqlConfig) string {
	return fmt.Sprintf(mysqlConnectionString, cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
}
