package main

import (
	"fmt"
	"log"
	"template-go-with-silverinha-file-genarator/internal/app"
	"template-go-with-silverinha-file-genarator/internal/infra/logger"
	"template-go-with-silverinha-file-genarator/internal/infra/utils"
	"template-go-with-silverinha-file-genarator/internal/infra/variables"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal(fmt.Sprintf("Error loading .env file. Err: %s", err.Error()))
	}
	utils.ShowBanner()
	logger.Init(&logger.Option{
		ServiceName:    variables.ServiceName(),
		ServiceVersion: variables.ServiceVersion(),
		Environment:    variables.Environment(),
		LogLevel:       variables.LogLevel(),
	})

	defer logger.Sync()

	application := app.Instance()
	application.Start(nil, false)
}
