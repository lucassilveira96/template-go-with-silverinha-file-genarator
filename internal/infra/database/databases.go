package database

import (
	"fmt"
	oracle "github.com/godoes/gorm-oracle"
	"gorm.io/gorm/schema"
	"runtime"
	"sync"
	"template-go-with-silverinha-file-genarator/internal/infra/variables"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const BIND_PARAM string = "_bind_"

type Databases struct {
	Read  *gorm.DB
	Write *gorm.DB
	Redis *Redis
}

func NewDatabases(c *fiber.Ctx) *Databases {
	dbs := &Databases{}
	var waitGroup sync.WaitGroup
	defer waitGroup.Wait()

	// Inicializar as conexões com os bancos
	dbs.buildReadDatabase(&waitGroup)
	dbs.buildWriteDatabase(&waitGroup)
	dbs.buildRedisDatabase(c, &waitGroup)

	return dbs
}

func (d *Databases) Close() {
	// Fechar as conexões do GORM
	if d.Read != nil {
		d.Read.Session(&gorm.Session{DryRun: true})
	}
	if d.Write != nil {
		d.Write.Session(&gorm.Session{DryRun: true})
	}
}

func (d *Databases) buildReadDatabase(waitGroup *sync.WaitGroup) {
	lazyConnection := variables.DBLazyConnection()
	cfg := &SqlConfig{
		ConnectionName:        variables.ServiceName() + "-read",
		Host:                  variables.DBHost(),
		Port:                  variables.DBPort(),
		Database:              variables.DBName(),
		Username:              variables.DBUsername(),
		Password:              variables.DBPassword(),
		MinConnections:        variables.DBMinConnections(),
		MaxConnections:        variables.DBMaxConnections(),
		ConnectionMaxLifetime: variables.DBConnectionMaxLifeTime(),
		ConnectionMaxIdleTime: variables.DBConnectionMaxIdleTime(),
		LazyConnection:        lazyConnection,
	}

	// Usando GORM para criar a conexão
	if lazyConnection {
		d.Read = d.connectDatabase(cfg)
	} else {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			d.Read = d.connectDatabase(cfg)
		}()
	}
}

func (d *Databases) buildWriteDatabase(waitGroup *sync.WaitGroup) {
	lazyConnection := variables.DBLazyConnection()
	cfg := &SqlConfig{
		ConnectionName:        variables.ServiceName() + "-write",
		Host:                  variables.DBHost(),
		Port:                  variables.DBPort(),
		Database:              variables.DBName(),
		Username:              variables.DBUsername(),
		Password:              variables.DBPassword(),
		MinConnections:        variables.DBMinConnections(),
		MaxConnections:        variables.DBMaxConnections(),
		ConnectionMaxLifetime: variables.DBConnectionMaxLifeTime(),
		ConnectionMaxIdleTime: variables.DBConnectionMaxIdleTime(),
		LazyConnection:        lazyConnection,
	}

	// Usando GORM para criar a conexão
	if lazyConnection {
		d.Write = d.connectDatabase(cfg)
	} else {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			d.Write = d.connectDatabase(cfg)
		}()
	}
}

func (d *Databases) buildRedisDatabase(c *fiber.Ctx, waitGroup *sync.WaitGroup) {
	lazyConnection := variables.RedisLazyConnection()
	opt := &redis.Options{
		Addr:         fmt.Sprintf("%s:%d", variables.RedisHost(), variables.RedisPort()),
		Password:     variables.RedisPassword(),
		DB:           variables.RedisDB(),
		PoolSize:     10 * runtime.NumCPU(),
		MinIdleConns: 10,
	}

	// Configuração para o Redis
	if lazyConnection {
		d.Redis = NewRedis(c, opt, lazyConnection)
	} else {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			d.Redis = NewRedis(c, opt, lazyConnection)
		}()
	}
}

func (d *Databases) connectDatabase(cfg *SqlConfig) *gorm.DB {
	var dsn string
	var db *gorm.DB
	var err error

	driver := variables.DatabaseType()

	switch driver {
	case "postgres":
		dsn = postgresConnectionStringBuilder(cfg)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	case "mysql":
		dsn = mariaDBConnectionStringBuilder(cfg)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	case "mariadb":
		dsn = mariaDBConnectionStringBuilder(cfg)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	case "oracle":
		dsn = oracleConnectionStringBuilder(cfg)
		dialector := oracle.New(oracle.Config{
			DSN:                       dsn,
			IgnoreCase:                false,
			NamingCaseSensitive:       true,
			VarcharSizeIsCharLength:   true,
			RowNumberAliasForOracle11: "ROW_NUM",
		})
		db, err = gorm.Open(dialector, &gorm.Config{
			SkipDefaultTransaction:                   true,
			DisableForeignKeyConstraintWhenMigrating: true,
			NamingStrategy: schema.NamingStrategy{
				NoLowerCase:         true,
				IdentifierMaxLength: 30,
			},
			PrepareStmt:     false,
			CreateBatchSize: 50,
		})
	default:
		panic(fmt.Sprintf("Unsupported database driver: %s", driver))
	}

	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database [%s]: %s", driver, err.Error()))
	}

	if variables.MigrationsEnabled() {
		d.runMigrations(db)
	}

	return db
}

func (d *Databases) runMigrations(db *gorm.DB) {
	err := db.AutoMigrate()
	if err != nil {
		panic(fmt.Sprintf("Failed to run migrations: %s", err.Error()))
	}
}
