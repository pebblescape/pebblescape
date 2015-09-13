package http

import (
	"log"
	"net/http"
	"time"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/ant0ine/go-json-rest/rest"
	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/stretchr/graceful"
	"github.com/pebblescape/pebblescape/host/gitreceive"
	"github.com/pebblescape/pebblescape/host/state"
)

const (
	LogFormat = "%S \033[0m\033[36;1m%Ts\033[0m \"%r\" \033[1;30m%u \"%{User-Agent}i\"\033[0m"
)

var DefaultStack = []rest.Middleware{
	&RequestIdMiddleware{},
	&rest.TimerMiddleware{},
	&rest.RecorderMiddleware{},
	&rest.RecoverMiddleware{},
}

func Serve(port string, s *state.State, repopath string, logger *log.Logger) {
	DefaultStack = append(
		[]rest.Middleware{
			&rest.AccessLogApacheMiddleware{
				Logger: logger,
				Format: LogFormat,
			},
		}, DefaultStack...)

	setupGit(s, repopath)
	setupHttp(s)

	// Create server
	server := &graceful.Server{
		Timeout: 10 * time.Second,
		Server: &http.Server{
			Addr: ":" + port,
		},
	}

	log.Println("HTTP API listening on " + port)
	log.Fatal(server.ListenAndServe())
}

func setupHttp(s *state.State) {
	api := rest.NewApi()
	handler := Api{s}

	router, err := rest.MakeRouter(
		rest.Get("/user", handler.ListUsers),
		rest.Get("/user/#user", handler.GetUser),
		rest.Post("/user", handler.AddUser),
		rest.Get("/app", handler.ListApps),
		rest.Get("/app/#app", handler.GetApp),
	)
	if err != nil {
		log.Fatal(err)
	}

	api.Use(DefaultStack...)
	api.Use(&rest.ContentTypeCheckerMiddleware{})
	api.SetApp(router)

	http.Handle("/", api.MakeHandler())
}

func setupGit(s *state.State, repopath string) {
	api := rest.NewApi()
	handler := gitreceive.NewGitHandler(repopath)

	api.Use(DefaultStack...)
	api.Use(&rest.AuthBasicMiddleware{
		Realm:         "Pebblescape GIT",
		Authenticator: s.Authenticate,
	})
	api.SetApp(rest.AppSimple(func(w rest.ResponseWriter, r *rest.Request) {
		handler.ServeHTTP(w.(http.ResponseWriter), r.Request)
	}))

	http.Handle("/git/", api.MakeHandler())
}
