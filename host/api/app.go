package api

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type App struct {
	Name string `json:"name,omitempty"`
}

func ParseApp(path string) (*App, error) {
	app := &App{}

	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(file, app); err != nil {
		return nil, err
	}

	return app, nil
}

func PersistApp(path string, app *App) error {
	bytes, err := json.Marshal(app)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(path, bytes, os.ModeExclusive|0600); err != nil {
		return err
	}

	return nil
}
