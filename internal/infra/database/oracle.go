package database

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

const (
	oracleDriver = "godror"
)

func NewOracle(c *fiber.Ctx, cfg *SqlConfig) *Database {
	if len(cfg.Driver) == 0 {
		cfg.Driver = oracleDriver
	}

	return NewDatabase(c, cfg, oracleConnectionStringBuilder)
}

func oracleConnectionStringBuilder(cfg *SqlConfig) string {
	return fmt.Sprintf(
		"user=\"%s\" password=\"%s\" connectString=\"%s:%s/%s\"",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database,
	)
}
