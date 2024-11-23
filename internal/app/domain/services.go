package domain

import (
	"template-go-with-silverinha-file-genarator/internal/infra/database"
)

type Services struct {
}

func NewServices(dbs *database.Databases) *Services {

	services := &Services{}
	return services
}
