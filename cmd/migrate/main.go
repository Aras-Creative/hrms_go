// Command migrate applies the SQL files in ./migrations without needing the
// golang-migrate CLI installed. The files are embedded into the binary, so a
// deployment is a single artifact with no external tooling step.
//
// Usage:
//
//	go run ./cmd/migrate up             apply every pending migration
//	go run ./cmd/migrate down [N]       roll back N migrations (default 1)
//	go run ./cmd/migrate down all       roll everything back to version 0
//	go run ./cmd/migrate version        print the current version and dirty flag
//	go run ./cmd/migrate force V        set the version without running anything
//	go run ./cmd/migrate drop          drop the version bookkeeping only
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"hrms/internal/pkg/config"
	"hrms/internal/pkg/database"
	"hrms/migrations"
)

// errNoChange marks the "already up to date" outcome so the process can exit 0
// without printing it as a failure.
var errNoChange = errors.New("no change")

func main() {
	if err := run(); err != nil && !errors.Is(err, errNoChange) {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	configPath := flag.String("config", "config/config.yaml", "path to the config file holding the database DSN")
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		usage()
		return errors.New("a command is required")
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("load config %s: %w", *configPath, err)
	}

	m, closeFn, err := newMigrator(cfg)
	if err != nil {
		return err
	}
	defer closeFn()

	switch args[0] {
	case "up":
		return report(m.Up())

	case "down":
		if len(args) == 2 && args[1] == "all" {
			fmt.Fprintln(os.Stderr, "warning: rolling back every migration drops all schema in this database")
			return report(m.Down())
		}
		steps := 1
		if len(args) == 2 {
			if _, err := fmt.Sscanf(args[1], "%d", &steps); err != nil || steps < 1 {
				return fmt.Errorf(`down expects a positive step count or "all", got %q`, args[1])
			}
		}
		return report(m.Steps(-steps))

	case "version":
		v, dirty, err := m.Version()
		if errors.Is(err, migrate.ErrNilVersion) {
			fmt.Println("no migration has been applied yet")
			return nil
		}
		if err != nil {
			return fmt.Errorf("read version: %w", err)
		}
		fmt.Printf("version=%d dirty=%v\n", v, dirty)
		if dirty {
			fmt.Fprintln(os.Stderr, "note: dirty means a migration failed midway; inspect the schema before forcing a version")
		}
		return nil

	case "force":
		if len(args) != 2 {
			return errors.New(`force expects a version, e.g. "force 47"`)
		}
		var v uint
		if _, err := fmt.Sscanf(args[1], "%d", &v); err != nil {
			return fmt.Errorf("force expects a numeric version, got %q", args[1])
		}
		if err := m.Force(int(v)); err != nil {
			return fmt.Errorf("force version %d: %w", v, err)
		}
		fmt.Printf("forced version to %d without running any migration\n", v)
		return nil

	case "drop":
		fmt.Fprintln(os.Stderr, "warning: drop removes only the version bookkeeping; the schema itself stays")
		if err := m.Drop(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("drop: %w", err)
		}
		fmt.Println("migration bookkeeping dropped")
		return nil

	default:
		usage()
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func newMigrator(cfg *config.Config) (*migrate.Migrate, func(), error) {
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return nil, nil, fmt.Errorf("read embedded migrations: %w", err)
	}

	// Reuse the app's own DSN builder so the CLI and the server can never drift.
	db, err := database.NewPostgres(&cfg.Database)
	if err != nil {
		return nil, nil, err
	}

	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("postgres driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("build migrator: %w", err)
	}
	return m, func() {
		_, _ = m.Close()
		_ = db.Close()
	}, nil
}

// report swallows ErrNoChange so re-running "up" on a current database is a
// quiet no-op instead of a non-zero exit.
func report(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, migrate.ErrNoChange):
		fmt.Println("database is already up to date")
		return errNoChange
	default:
		return err
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: migrate [-config path] <up|down [N|all]|version|force V|drop>")
	fmt.Fprintln(os.Stderr, "  up           apply every pending migration")
	fmt.Fprintln(os.Stderr, "  down [N]     roll back N migrations (default 1, \"all\" rolls back everything)")
	fmt.Fprintln(os.Stderr, "  version      print the current version and dirty flag")
	fmt.Fprintln(os.Stderr, "  force V      set the recorded version without running any migration")
	fmt.Fprintln(os.Stderr, "  drop         remove the version bookkeeping (leaves the schema in place)")
}
