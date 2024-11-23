package adapters

import (
	"github.com/gofiber/fiber/v2"
)

type Handlers struct {
}

func NewHandlers(services *domain.Services) *Handlers {
	return &Handlers{}
}

func (h *Handlers) Configure(server *fiber.App) {
}
