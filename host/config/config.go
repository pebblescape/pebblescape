package config

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pebblescape/pebblescape/pkg/random"
)

func Open(file string) (*Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}

func Parse(r io.Reader) (*Config, error) {
	conf := &Config{}
	if err := json.NewDecoder(r).Decode(conf); err != nil {
		return nil, err
	}
	return conf, nil
}

type Config struct {
	HostKey string `json:"host_key,omitempty"`
	Home    string `json:"home,omitempty"`
	DbID    string `json:"db_id,omitempty"`
	DbPass  string `json:"db_oass,omitempty"`
	File    string `json:"-"`
}

func New() *Config {
	return &Config{}
}

func Ensure(name, key, home string) (*Config, error) {
	c := New()

	_, err := os.OpenFile(name, os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return c, err
	}

	home, err = filepath.Abs(home)

	c.File = name
	c.Home = home
	c.HostKey = random.Hex(20)

	if key != "" {
		c.HostKey = key
	}

	if err := c.Write(); err != nil {
		return c, err
	}

	return c, nil
}

func (c *Config) EnsurePaths() error {
	if err := os.MkdirAll(filepath.Join(c.Home, "db"), 0770); !os.IsExist(err) {
		return err
	}

	return nil
}

func (c *Config) Write() error {
	data, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return err
	}

	file, err := ioutil.TempFile("", "pebblescape-conf")
	if err != nil {
		return err
	}

	file.Write(append(data, '\n'))
	if err := file.Sync(); err != nil {
		return err
	}

	file.Close()

	if err := os.Rename(file.Name(), c.File); err != nil {
		os.Remove(file.Name())
		return err
	}

	return nil
}

func (c *Config) Cleanup() error {
	return os.Remove(c.File)
}
