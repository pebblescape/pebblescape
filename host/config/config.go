// Package config manages everything related to the config file of a running host instance.
package config

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pebblescape/pebblescape/pkg/random"
)

// ConfigFile specifies default config file location.
const ConfigFile = "/var/run/pebblescape.json"

// Open parses config file at specifided location.
func Open(file string) (*Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}

// Parse decodes the json encoded the config file.
func Parse(r io.Reader) (*Config, error) {
	conf := &Config{}
	if err := json.NewDecoder(r).Decode(conf); err != nil {
		return nil, err
	}
	return conf, nil
}

// Config holds all information about a running host instance.
type Config struct {
	PID     string `json:"pid,omitempty"`
	HostKey string `json:"host_key,omitempty"`
	Home    string `json:"home,omitempty"`
	DbID    string `json:"db_id,omitempty"`
	DbPass  string `json:"db_pass,omitempty"`
	File    string `json:"-"`
}

// New creates new Config.
func New() *Config {
	return &Config{}
}

// Ensure opens and writes a config file in exclusive mode ensuring that only one host instance is running.
func Ensure(name, key, home string) (*Config, error) {
	c := New()

	if err := c.EnsurePaths(); err != nil {
		return c, err
	}

	_, err := os.OpenFile(name, os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return c, err
	}

	home, err = filepath.Abs(home)

	c.PID = string(os.Getpid())
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

// EnsurePaths ensures that all paths necessary for proper host functioning exist.
func (c *Config) EnsurePaths() error {
	if err := os.MkdirAll(filepath.Dir(ConfigFile), 0770); !os.IsExist(err) {
		return err
	}

	if err := os.MkdirAll(filepath.Join(c.Home, "db"), 0770); !os.IsExist(err) {
		return err
	}

	return nil
}

// Write serializes and writes current configuration atomically.
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

// Cleanup removes config file of current instance.
func (c *Config) Cleanup() error {
	return os.Remove(c.File)
}
