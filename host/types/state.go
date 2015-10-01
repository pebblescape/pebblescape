package host

type State interface {
	Restore(backend Backend) (func(), error)
	Authenticate(username string, password string) bool
	ListApps() map[string]App
	GetApp(name string) *App
	ListJobs() []*Job
	GetJob(id string) *Job
	RemoveJob(id string)
	ListUsers() []*User
	GetUser(name string) *User
	AddUser(user *User) error
	AddJob(j *Job)
	RunJob(j *Job) error
	SetContainerID(jobID, containerID string)
}
