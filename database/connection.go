package database

// TODO ENGINEER: migrate from sqlx to pgx
import (
	"time"

	"github.com/yisaj/heavens_throne/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	dbDriverName          = "postgres"
	maxConnectionAttempts = 10
	migrationsURL         = "file://migrations"
)

// Resource is the wrapper interface that includes all database access methods
type Resource interface {
	LocationResource
	PlayerResource
	WebhooksResource
	GameResource
}

type connection struct {
	db *sqlx.DB
}

// Connect opens a connection to the database and returns the resource object
func Connect(conf *config.Config, logger *logrus.Logger) (Resource, error) {
	// open and ping db connection
	var db *sqlx.DB
	var err error
	for attempts := 1; attempts <= maxConnectionAttempts; attempts++ {
		db, err = sqlx.Connect(dbDriverName, conf.DatabaseURI)
		if err == nil {
			break
		}

		logger.Errorf("Database connection error: %s\n", err.Error())
		time.Sleep(time.Second)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed database connection")
	}

	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return nil, errors.Wrap(err, "failed database driver construction")
	}

	migration, err := migrate.NewWithDatabaseInstance(migrationsURL, dbDriverName, driver)
	if err != nil {
		return nil, errors.Wrap(err, "failed database migrator construction")
	}

	err = migration.Up()
	if err != nil && err != migrate.ErrNoChange {
		return nil, errors.Wrap(err, "failed database migration")
	}

	return &connection{db}, nil
}
