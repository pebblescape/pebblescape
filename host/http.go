package main

import (
	"log"
	"net"
	"net/http"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/julienschmidt/httprouter"
	"github.com/pebblescape/pebblescape/pkg/httphelper"
	"github.com/pebblescape/pebblescape/pkg/shutdown"
)

type httpAPI struct {
	state *State
}

func (h *httpAPI) ListUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	httphelper.JSON(w, 200, h.state.GetUsers())
}

func (h *httpAPI) RegisterRoutes(r *httprouter.Router) error {
	r.GET("/user", h.ListUsers)
	return nil
}

func serveHTTP(s *State) error {
	listener, err := net.Listen("tcp", ":4592")
	if err != nil {
		return err
	}
	shutdown.BeforeExit(func() { listener.Close() })
	log.Println("Listening for HTTP connections on port 4592")

	router := httprouter.New()
	httpAPI := &httpAPI{
		state: s,
	}
	httpAPI.RegisterRoutes(router)

	go http.Serve(listener, router)

	return nil
}
