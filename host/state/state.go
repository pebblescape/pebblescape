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
	"github.com/pebblescape/pebblescape/host/types"
	"github.com/pebblescape/pebblescape/pkg/random"
)

type State struct {
	jobs  map[string]*host.Job
	users map[string]*host.User
	apps  map[string]*host.App

	mtx           sync.RWMutex
	stateFilePath string
	stateDB       *bolt.DB

	backend host.Backend
}

func NewState(stateFilePath string) host.State {
	s := &State{
		stateFilePath: stateFilePath,
		jobs:          make(map[string]*host.Job),
		users:         make(map[string]*host.User),
		apps:          make(map[string]*host.App),
	}
	s.initializePersistence()
	return s
}

func (s *State) Restore(backend host.Backend) (func(), error) {
	s.backend = backend

	var resurrect []*host.Job
	if err := s.stateDB.View(func(tx *bolt.Tx) error {
		usersBucket := tx.Bucket([]byte("users"))
		jobsBucket := tx.Bucket([]byte("jobs"))

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

		// restore jobs
		if err := jobsBucket.ForEach(func(k, v []byte) error {
			job := &host.Job{}
			if err := json.Unmarshal(v, job); err != nil {
				return err
			}

			s.jobs[string(k)] = job
			resurrect = append(resurrect, job)

			return nil
		}); err != nil {
			return err
		}

		return nil
	}); err != nil && err != io.EOF {
		return nil, fmt.Errorf("Could not restore from host persistence db: %s", err)
	}

	return func() {
		var wg sync.WaitGroup
		wg.Add(len(resurrect))
		for _, job := range resurrect {
			go func(job *host.Job) {
				backend.Restore(job)
				wg.Done()
			}(job)
		}
		wg.Wait()
	}, nil
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

func (s *State) ListJobs() []*host.Job {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	res := make([]*host.Job, 0)
	for _, v := range s.jobs {
		res = append(res, v)
	}
	return res
}

func (s *State) GetJob(name string) *host.Job {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	job := s.jobs[name]
	if job == nil {
		return nil
	}
	jobCopy := *job
	return &jobCopy
}

func (s *State) RunJob(j *host.Job) error {
	s.AddJob(j)
	return s.backend.Run(j)
}

func (s *State) AddJob(j *host.Job) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if j.ID == "" {
		j.ID = random.Hex(10)
	}
	s.jobs[j.ID] = j
	s.persistJob(j.ID)
}

func (s *State) RemoveJob(id string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	delete(s.jobs, id)
	s.persistJob(id)
}

func (s *State) SetContainerID(jobID, containerID string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.jobs[jobID].ContainerID = containerID
	s.persistJob(jobID)
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

	existing := s.users[user.Name]
	if existing != nil {
		return fmt.Errorf("User already exists")
	}

	if err := user.Create(); err != nil {
		return err
	}

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
		panic(fmt.Errorf("Could not not mkdir for db: %s", err))
	}
	stateDB, err := bolt.Open(s.stateFilePath, 0600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		panic(fmt.Errorf("Could not open db: %s", err))
	}
	s.stateDB = stateDB
	if err := s.stateDB.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte("jobs"))
		tx.CreateBucketIfNotExists([]byte("users"))
		tx.CreateBucketIfNotExists([]byte("apps"))
		return nil
	}); err != nil {
		panic(fmt.Errorf("Could not initialize host persistence db: %s", err))
	}
}

func (s *State) persistUser(username string) {
	if err := s.stateDB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("users"))

		if _, exists := s.users[username]; exists {
			b, err := json.Marshal(s.users[username])
			if err != nil {
				return fmt.Errorf("Failed to serialize user: %s", err)
			}
			err = bucket.Put([]byte(username), b)
			if err != nil {
				return fmt.Errorf("Could not persist user to db: %s", err)
			}
		} else {
			bucket.Delete([]byte(username))
		}

		return nil
	}); err != nil {
		panic(fmt.Errorf("Could not persist user to db: %s", err))
	}
}

func (s *State) persistJob(jobID string) {
	if err := s.stateDB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("jobs"))

		if _, exists := s.jobs[jobID]; exists {
			b, err := json.Marshal(s.jobs[jobID])
			if err != nil {
				return fmt.Errorf("Failed to serialize job: %s", err)
			}
			err = bucket.Put([]byte(jobID), b)
			if err != nil {
				return fmt.Errorf("Could not persist job to db: %s", err)
			}
		} else {
			bucket.Delete([]byte(jobID))
		}

		return nil
	}); err != nil {
		panic(fmt.Errorf("Could not persist job to db: %s", err))
	}
}
