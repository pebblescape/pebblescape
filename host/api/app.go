package api

import (
	"database/sql"
	"errors"
	"time"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/jmoiron/sqlx"
	"github.com/pebblescape/pebblescape/pkg/utils"
)

var (
	ErrAppNameInvalid = errors.New("App name contains invalid characters. Alphanumeric only, no whitespace.")
	ErrAppNotExist    = errors.New("App does not exist")
	ErrAppNameTooLong = errors.New("App name too long. Max characters 30.")
	AppNameMax        = 30
)

// App represents the app in the database.
type App struct {
	ID        string     `db:"id"`
	Name      string     `db:"name"`
	CreatedAt *time.Time `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

// AppRepo handles all interfacing between the host API and database.
type AppRepo struct {
	API *API
	DB  *sqlx.DB
}

// GetAppRepo creates and returns a new AppRepo instance.
func (a *API) GetAppRepo() *AppRepo {
	return &AppRepo{a, a.DB}
}

// List returns all apps in the database.
func (r *AppRepo) List() ([]App, error) {
	apps := []App{}

	if err := r.DB.Select(&apps, "SELECT * FROM apps ORDER BY name ASC"); err != nil {
		return apps, err
	}

	return apps, nil
}

// Get finds an app by ID and returns it.
func (r *AppRepo) Get(id string) (*App, error) {
	app := &App{}

	if err := r.DB.Get(app, "SELECT * FROM apps WHERE id=$1", id); err != nil {
		if err == sql.ErrNoRows {
			return app, ErrAppNotExist
		}
		return app, err
	}

	return app, nil
}

// GetByName finds an app by name and returns it.
func (r *AppRepo) GetByName(name string) (*App, error) {
	app := &App{}

	if err := r.DB.Get(app, "SELECT * FROM apps WHERE name=$1", name); err != nil {
		if err == sql.ErrNoRows {
			return app, ErrAppNotExist
		}
		return app, err
	}

	return app, nil
}

// Create saves a new app into the database.
func (r *AppRepo) Create(app *App) error {
	if !utils.AppNamePattern.Match([]byte(app.Name)) {
		return ErrAppNameInvalid
	}

	if len(app.Name) > AppNameMax {
		return ErrAppNameTooLong
	}

	if _, err := r.DB.NamedExec(`INSERT INTO apps (name) VALUES (:name)`, app); err != nil {
		return err
	}

	newApp, err := r.GetByName(app.Name)
	if err != nil {
		return err
	}

	app.ID = newApp.ID
	app.CreatedAt = newApp.CreatedAt
	app.UpdatedAt = newApp.UpdatedAt

	return nil
}

// Update changes values of an existing app.
func (r *AppRepo) Update(id string, data map[string]interface{}) (*App, error) {
	app, err := r.Get(id)
	if err != nil {
		return nil, err
	}

	return app, nil
}
