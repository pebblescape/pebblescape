package state

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/boltdb/bolt"
	"github.com/pebblescape/pebblescape/host/backend"
	"github.com/pebblescape/pebblescape/host/types"
)

type State struct {
	jobs  map[string]*host.Job
	users map[string]*host.User
	apps  map[string]*host.App

	mtx           sync.RWMutex
	stateFilePath string
	stateDB       *bolt.DB

	backend backend.Backend
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

func (s *State) Restore(backend backend.Backend) error {
	s.backend = backend

	if err := s.stateDB.View(func(tx *bolt.Tx) error {
		usersBucket := tx.Bucket([]byte("users"))

		// restore users
		if err := usersBucket.ForEach(func(k, v []byte) error {
			user := &host.User{}
			if err := json.Unmarshal(v, user); err != nil {
				return err
			}

			s.users[string(k)] = user
			return nil
		}); err != nil {
			return err
		}

		testuser := &host.User{Name: "yoooo"}
		s.users[testuser.Name] = testuser

		return nil
	}); err != nil && err != io.EOF {
		return fmt.Errorf("could not restore from host persistence db: %s", err)
	}

	return nil
}

func (s *State) Authenticate(username string, password string) bool {
	user := s.GetUser(username)
	return user != nil
}

func (s *State) ListApps() map[string]host.App {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	res := make(map[string]host.App, len(s.apps))
	for k, v := range s.apps {
		res[k] = *v
	}
	return res
}

func (s *State) GetApp(name string) *host.App {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	app := s.apps[name]
	if app == nil {
		return nil
	}
	appCopy := *app
	return &appCopy
}

func (s *State) ListUsers() []*host.User {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	res := make([]*host.User, 0)
	for _, v := range s.users {
		res = append(res, v)
	}
	return res
}

func (s *State) GetUser(name string) *host.User {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	user := s.users[name]
	if user == nil {
		return nil
	}
	userCopy := *user
	return &userCopy
}

func (s *State) AddUser(user *host.User) error {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	s.users[user.Name] = user
	s.persistUser(user.Name)
	return nil
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
		tx.CreateBucketIfNotExists([]byte("users"))
		tx.CreateBucketIfNotExists([]byte("apps"))
		return nil
	}); err != nil {
		panic(fmt.Errorf("could not initialize host persistence db: %s", err))
	}
}

func (s *State) persistUser(username string) {
	if err := s.stateDB.Update(func(tx *bolt.Tx) error {
		userBucket := tx.Bucket([]byte("users"))

		if _, exists := s.users[username]; exists {
			b, err := json.Marshal(s.users[username])
			if err != nil {
				return fmt.Errorf("failed to serialize user: %s", err)
			}
			err = userBucket.Put([]byte(username), b)
			if err != nil {
				return fmt.Errorf("could not persist user to boltdb: %s", err)
			}
		} else {
			userBucket.Delete([]byte(username))
		}

		return nil
	}); err != nil {
		panic(fmt.Errorf("could not persist to boltdb: %s", err))
	}
}
