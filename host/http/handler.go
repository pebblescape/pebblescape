package http

import (
	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/ant0ine/go-json-rest/rest"
	"github.com/pebblescape/pebblescape/host/api"
)

type httpHandler struct {
	API *api.API
}

func (h *httpHandler) ListJobs(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(nil)
}
