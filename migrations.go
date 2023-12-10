package medego

import (
	"log"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func (c *Medego) runMigration(dsn string, operation func(*migrate.Migrate) error) error {
	rootPath := filepath.ToSlash(c.RootPath)
	m, err := migrate.New("file://"+rootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := operation(m); err != nil {
		log.Println("Error running migration:", err)
		return err
	}

	return nil
}

func (c *Medego) MigrateUp(dsn string) error {
	return c.runMigration(dsn, (*migrate.Migrate).Up)
}

func (c *Medego) MigrateDownAll(dsn string) error {
	return c.runMigration(dsn, (*migrate.Migrate).Down)
}

func (c *Medego) Steps(n int, dsn string) error {
	return c.runMigration(dsn, func(m *migrate.Migrate) error {
		return m.Steps(n)
	})
}

func (c *Medego) MigrateForce(dsn string) error {
	return c.runMigration(dsn, func(m *migrate.Migrate) error {
		return m.Force(-1)
	})
}
