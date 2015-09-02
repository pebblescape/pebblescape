package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/boltdb/bolt"
	"github.com/pebblescape/pebblescape/host/types"
)

type State struct {
	jobs  map[string]*host.Job
	users map[string]*host.User
	apps  map[string]*host.App

	mtx           sync.RWMutex
	stateFilePath string
	stateDB       *bolt.DB

	backend Backend
}

func NewState(stateFilePath string) *State {
	s := &State{
		stateFilePath: stateFilePath,
		jobs:          make(map[string]*host.Job),
		users:         make(map[string]*host.User),
		apps:          make(map[string]*host.App),
	}
	s.initializePersistence()
	return s
}

func (s *State) Restore(backend Backend) {
	s.backend = backend
}

func (s *State) GetApp(name string) host.App {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	app := *s.apps[name]
	return app
}

func (s *State) GetUsers() map[string]host.User {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	res := make(map[string]host.User, len(s.users))
	for k, v := range s.users {
		res[k] = *v
	}
	return res
}

func (s *State) initializePersistence() {
	if s.stateDB != nil {
		return
	}

	// open/initialize db
	if err := os.MkdirAll(filepath.Dir(s.stateFilePath), 0755); err != nil {
		panic(fmt.Errorf("could not not mkdir for db: %s", err))
	}
	stateDB, err := bolt.Open(s.stateFilePath, 0600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		panic(fmt.Errorf("could not open db: %s", err))
	}
	s.stateDB = stateDB
	if err := s.stateDB.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte("jobs"))
		return nil
	}); err != nil {
		panic(fmt.Errorf("could not initialize host persistence db: %s", err))
	}
}
