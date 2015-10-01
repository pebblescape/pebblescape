package http

import (
	"net/http"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/ant0ine/go-json-rest/rest"
	"github.com/pebblescape/pebblescape/host/types"
)

type Api struct {
	state host.State
}

func (a *Api) ListApps(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(a.state.ListApps())
}

func (a *Api) GetApp(w rest.ResponseWriter, r *rest.Request) {
	app := a.state.GetApp(r.PathParams["app"])
	if app == nil {
		rest.NotFound(w, r)
		return
	}
	w.WriteJson(app)
}

func (a *Api) ListJobs(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(a.state.ListJobs())
}

func (a *Api) GetJob(w rest.ResponseWriter, r *rest.Request) {
	job := a.state.GetJob(r.PathParams["job"])
	if job == nil {
		rest.NotFound(w, r)
		return
	}
	w.WriteJson(job)
}

func (a *Api) ListUsers(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(a.state.ListUsers())
}

func (a *Api) GetUser(w rest.ResponseWriter, r *rest.Request) {
	user := a.state.GetUser(r.PathParams["user"])
	if user == nil {
		rest.NotFound(w, r)
		return
	}
	w.WriteJson(user)
}

func (a *Api) AddUser(w rest.ResponseWriter, r *rest.Request) {
	user := host.User{}
	err := r.DecodeJsonPayload(&user)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = a.state.AddUser(&user)
	if err != nil {
		rest.Error(w, err.Error(), 400)
		return
	}
	w.WriteJson(&user)
}
