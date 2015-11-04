package api

import (
	"database/sql"
	"errors"
	"time"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/jmoiron/sqlx"
	"github.com/pebblescape/pebblescape/pkg/utils"
)

var (
	AppNameError   = errors.New("App name contains invalid characters. Alphanumeric only, no whitespace.")
	AppNotExist    = errors.New("App does not exist")
	AppNameTooLong = errors.New("App name too long. Max characters 30.")
	AppNameMax     = 30
)

type App struct {
	ID        string     `db:"id"`
	Name      string     `db:"name"`
	CreatedAt *time.Time `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

type AppRepo struct {
	Api *Api
	DB  *sqlx.DB
}

func (a *Api) GetAppRepo() *AppRepo {
	return &AppRepo{a, a.DB}
}

func (r *AppRepo) List() ([]App, error) {
	apps := []App{}

	if err := r.DB.Select(&apps, "SELECT * FROM apps ORDER BY name ASC"); err != nil {
		return apps, err
	}

	return apps, nil
}

func (r *AppRepo) Get(id string) (*App, error) {
	app := &App{}

	if err := r.DB.Get(app, "SELECT * FROM apps WHERE id=$1", id); err != nil {
		if err == sql.ErrNoRows {
			return app, AppNotExist
		}
		return app, err
	}

	return app, nil
}

func (r *AppRepo) GetByName(name string) (*App, error) {
	app := &App{}

	if err := r.DB.Get(app, "SELECT * FROM apps WHERE name=$1", name); err != nil {
		if err == sql.ErrNoRows {
			return app, AppNotExist
		}
		return app, err
	}

	return app, nil
}

func (r *AppRepo) Create(app *App) error {
	if !utils.AppNamePattern.Match([]byte(app.Name)) {
		return AppNameError
	}

	if len(app.Name) > AppNameMax {
		return AppNameTooLong
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

func (r *AppRepo) Update(id string, data map[string]interface{}) (*App, error) {
	app, err := r.Get(id)
	if err != nil {
		return nil, err
	}

	return app, nil
}
