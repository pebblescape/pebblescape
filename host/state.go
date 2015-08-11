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
	jobs map[string]*host.Job

	mtx           sync.RWMutex
	stateFilePath string
	stateDB       *bolt.DB

	backend Backend
}

func NewState(stateFilePath string) *State {
	s := &State{
		stateFilePath: stateFilePath,
		jobs:          make(map[string]*host.Job),
	}
	s.initializePersistence()
	return s
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
