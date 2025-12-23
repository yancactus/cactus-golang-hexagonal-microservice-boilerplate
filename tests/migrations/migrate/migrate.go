package migrate

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // postgres driver for migrate
	_ "github.com/golang-migrate/migrate/v4/source/file"       // file source for migrate
	_ "github.com/lib/pq"                                      // postgres driver

	"cactus-golang-hexagonal-microservice-boilerplate/config"
)

func PostgreMigrateUp(conf *config.Config) error {
	if conf.MigrationDir == "" {
		return nil
	}

	m, err := migrate.New(
		"file://"+conf.MigrationDir,
		fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=%s",
			conf.Postgre.User,
			conf.Postgre.Password,
			conf.Postgre.Host,
			conf.Postgre.Port,
			conf.Postgre.Database,
			conf.Postgre.SSLMode,
		),
	)
	if err != nil {
		return err
	}
	defer func() { _, _ = m.Close() }()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func PostgreMigrateDrop(conf *config.Config) error {
	m, err := migrate.New(
		"file://"+conf.MigrationDir,
		fmt.Sprintf(
			"postgresql://%s:%s@%s:%d/%s?sslmode=%s",
			conf.Postgre.User,
			conf.Postgre.Password,
			conf.Postgre.Host,
			conf.Postgre.Port,
			conf.Postgre.Database,
			conf.Postgre.SSLMode,
		),
	)
	if err != nil {
		return err
	}
	defer func() { _, _ = m.Close() }()

	if err := m.Drop(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
