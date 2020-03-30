package database

import (
	"time"

	"github.com/yisaj/heavens_throne/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

const (
	dbDriverName          = "postgres"
	maxConnectionAttempts = 10
	migrationsURL         = "file://migrations"
)

type Resource interface {
	LocationResource
	PlayerResource
	WebhooksResource
	GameResource
}

type connection struct {
	db *sqlx.DB
}

func Connect(conf *config.Config) (Resource, error) {
	// open and ping db connection
	var db *sqlx.DB
	var err error
	for attempts := 1; attempts <= maxConnectionAttempts; attempts++ {
		db, err = sqlx.Connect(dbDriverName, conf.DatabaseURI)
		if err == nil {
			break
		}
		// TODO: log an error here
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
