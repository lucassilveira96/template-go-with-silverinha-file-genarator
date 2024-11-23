package app

import (
	"archive/zip"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	adapters "template-go-with-silverinha-file-genarator/internal/app/adapter"
	"template-go-with-silverinha-file-genarator/internal/app/domain"
	"template-go-with-silverinha-file-genarator/internal/infra/database"
	"template-go-with-silverinha-file-genarator/internal/infra/logger"
	"template-go-with-silverinha-file-genarator/internal/infra/logger/attributes"
	"template-go-with-silverinha-file-genarator/internal/infra/server"
	"template-go-with-silverinha-file-genarator/internal/infra/variables"
	"time"

	"github.com/gofiber/fiber/v2"
)

type App struct {
	running bool
	locker  sync.Mutex

	server    *fiber.App
	handlers  *adapters.Handlers
	services  *domain.Services
	databases *database.Databases
}

var app = new(App)

func Instance() *App {
	return app
}

func (app *App) Start(c *fiber.Ctx, async bool) {
	app.locker.Lock()

	if app.running {
		app.locker.Unlock()
		return
	}

	start := time.Now()

	app.build(c)

	// Inicia o backup do log às 00:00:00 do dia seguinte
	go app.scheduleBackupAtMidnight()

	if async {
		go app.startServer(c, start)
	} else {
		app.startServer(c, start)
	}
}

func (app *App) Stop(c *fiber.Ctx) {
	app.locker.Lock()

	if !app.running {
		app.locker.Unlock()
		return
	}

	defer app.setRunning(false)
	defer app.locker.Unlock()

	if err := app.server.Shutdown(); err != nil {
		logger.Error(
			c,
			"Erro ao tentar fechar o servidor Fiber",
			attributes.New().WithError(err),
		)
	}

	app.databases.Close()
	app.dispose()

	logger.Warn(c, "Aplicação parada", nil)
}

func (app *App) IsRunning() bool {
	return app.running
}

func (app *App) startServer(c *fiber.Ctx, start time.Time) {
	defer app.setRunning(false)

	go func() {
		app.locker.Unlock()
	}()

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		<-ch
		app.Stop(c)
	}()

	port := ":" + strconv.Itoa(variables.ServerPort())
	if err := app.server.Listen(port); err != nil {
		logger.Warn(c, "Aplicação parada de forma graciosa", attributes.New().WithError(err))
	}
}

func (app *App) build(c *fiber.Ctx) {
	app.databases = database.NewDatabases(c)
	app.services = domain.NewServices(app.databases)
	app.handlers = adapters.NewHandlers(app.services)
	app.server = server.New()
	app.handlers.Configure(app.server)
}

func (app *App) dispose() {
	app.server = nil
	app.handlers = nil
	app.services = nil
	app.databases = nil
}

func (app *App) setRunning(run bool) {
	app.running = run
}

func (app *App) getTimeUntilMidnight() time.Duration {
	now := time.Now()
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	return nextMidnight.Sub(now)
}

func (app *App) scheduleBackupAtMidnight() {
	timeUntilMidnight := app.getTimeUntilMidnight()

	time.Sleep(timeUntilMidnight)

	logger.Info(nil, "Iniciando backup dos logs", nil)
	app.backupLogs()

	go app.scheduleBackupAtMidnight()
}

func (app *App) compressLogFile(logFile string) (string, error) {
	backupZipFile := fmt.Sprintf("%s.zip", logFile)
	zipFile, err := os.Create(backupZipFile)
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	logFileContent, err := os.Open(logFile)
	if err != nil {
		return "", err
	}
	defer logFileContent.Close()

	zipEntry, err := zipWriter.Create(logFile)
	if err != nil {
		return "", err
	}

	_, err = fmt.Fprint(zipEntry, logFileContent)
	if err != nil {
		return "", err
	}

	return backupZipFile, nil
}

func (app *App) backupLogs() {
	logFile := fmt.Sprintf("%s-debug.log", time.Now().Format("02-01-2006"))
	backupDir := fmt.Sprintf("%sbackup/", variables.DirLog())

	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		err := os.Mkdir(backupDir, 0755)
		if err != nil {
			logger.Error(nil, "Falha ao criar diretório de backup", attributes.New().WithError(err))
			return
		}
	}

	backupZipFile, err := app.compressLogFile(fmt.Sprintf("%s%s", variables.DirLog(), logFile))
	if err != nil {
		logger.Error(nil, "Falha ao comprimir arquivo de log", attributes.New().WithError(err))
		return
	}

	backupFile := fmt.Sprintf("%s%s", backupDir, backupZipFile)
	err = os.Rename(backupZipFile, backupFile)
	if err != nil {
		logger.Error(nil, "Falha ao mover arquivo comprimido para o backup", attributes.New().WithError(err))
		return
	}

	logger.Info(nil, fmt.Sprintf("Backup de log comprimido realizado: %s", backupFile), nil)
}
