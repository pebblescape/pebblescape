package postgres

import (
	"database/sql"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/jmoiron/sqlx"
)

// Migration specifies the version and SQL of a single database migration.
type Migration struct {
	ID    int
	Stmts []string
}

// NewMigrations returns a new list migrations.
func NewMigrations() *Migrations {
	l := make(Migrations, 0)
	return &l
}

// Migrations is an array of migrations.
type Migrations []Migration

// Add appends a migration the list.
func (m *Migrations) Add(id int, stmts ...string) {
	*m = append(*m, Migration{ID: id, Stmts: stmts})
}

// Migrate executes all necessary migrations in the specified database.
func (m Migrations) Migrate(db *sqlx.DB) error {
	var initialized bool
	for _, migration := range m {
		if !initialized {
			db.Exec("CREATE TABLE IF NOT EXISTS schema_migrations (id bigint PRIMARY KEY)")
			initialized = true
		}

		tx, err := db.Begin()
		if err != nil {
			return err
		}

		if _, err := tx.Exec("LOCK TABLE schema_migrations IN ACCESS EXCLUSIVE MODE"); err != nil {
			tx.Rollback()
			return err
		}
		var tmp bool
		if err := tx.QueryRow("SELECT true FROM schema_migrations WHERE id = $1", migration.ID).Scan(&tmp); err != sql.ErrNoRows {
			tx.Rollback()
			if err == nil {
				continue
			}
			return err
		}

		for _, s := range migration.Stmts {
			if _, err = tx.Exec(s); err != nil {
				tx.Rollback()
				return err
			}
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (id) VALUES ($1)", migration.ID); err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
