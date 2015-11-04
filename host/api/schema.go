package api

import (
	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/jmoiron/sqlx"
	"github.com/pebblescape/pebblescape/pkg/postgres"
)

func migrateDB(db *sqlx.DB) error {
	m := postgres.NewMigrations()
	m.Add(1,
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,

		`CREATE TABLE apps (
			id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
			name text NOT NULL,
			created_at timestamptz NOT NULL DEFAULT now(),
			updated_at timestamptz NOT NULL DEFAULT now(),
			deleted_at timestamptz
		)`,
		`CREATE UNIQUE INDEX ON apps (name) WHERE deleted_at IS NULL`,
	)
	m.Add(2,
		`CREATE FUNCTION update_timestamp() RETURNS TRIGGER AS $$
		    BEGIN
		        NEW.updated_at = now();
   				RETURN NEW;
		    END;
		$$ LANGUAGE plpgsql`,

		`CREATE TRIGGER apps_timestamp_trigger
    		BEFORE UPDATE ON apps
    		FOR EACH ROW EXECUTE PROCEDURE update_timestamp()`,
	)

	return m.Migrate(db)
}
