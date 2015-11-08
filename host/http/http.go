package http

import (
	"log"
	"net"
	"net/http"
	"path/filepath"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/ant0ine/go-json-rest/rest"
	"github.com/pebblescape/pebblescape/host/api"
	"github.com/pebblescape/pebblescape/host/config"
	"github.com/pebblescape/pebblescape/host/gitreceive"
	"github.com/pebblescape/pebblescape/pkg/shutdown"
)

const (
	logFormat = "%S \033[0m\033[36;1m%Ts\033[0m \"%r\" \033[1;30m%u \"%{User-Agent}i\"\033[0m"
)

var defaultStack = []rest.Middleware{
	&requestIDMiddleware{},
	&rest.TimerMiddleware{},
	&rest.RecorderMiddleware{},
	&rest.RecoverMiddleware{},
}

// Serve starts HTTP and Git APIs.
func Serve(port string, hostAPI *api.API, conf *config.Config, logger *log.Logger) error {
	defaultStack = append(
		[]rest.Middleware{
			&rest.AccessLogApacheMiddleware{
				Logger: logger,
				Format: logFormat,
			},
		}, defaultStack...)

	auth := func(user string, password string) bool {
		return password == conf.HostKey
	}

	setupGit(auth, conf)
	setupHTTP(auth, hostAPI)

	// Create server
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	shutdown.BeforeExit(func() { listener.Close() })

	log.Println("HTTP API listening on " + port)
	return http.Serve(listener, nil)
}

func setupHTTP(auth func(string, string) bool, hostAPI *api.API) {
	restAPI := rest.NewApi()
	handler := httpHandler{hostAPI}

	router, err := rest.MakeRouter(
		rest.Get("/job", handler.ListJobs),
	)
	if err != nil {
		log.Fatal(err)
	}

	restAPI.Use(defaultStack...)
	restAPI.Use(&rest.AuthBasicMiddleware{
		Realm:         "Pebblescape",
		Authenticator: auth,
	})
	restAPI.Use(&rest.ContentTypeCheckerMiddleware{})
	restAPI.SetApp(router)

	http.Handle("/", restAPI.MakeHandler())
}

func setupGit(auth func(string, string) bool, conf *config.Config) {
	cachePath := filepath.Join(conf.Home, "cache")
	repoPath := filepath.Join(conf.Home, "repos")

	restAPI := rest.NewApi()
	handler := gitreceive.NewGitHandler(repoPath, cachePath)

	restAPI.Use(defaultStack...)
	restAPI.Use(&rest.AuthBasicMiddleware{
		Realm:         "Pebblescape GIT",
		Authenticator: auth,
	})
	restAPI.SetApp(rest.AppSimple(func(w rest.ResponseWriter, r *rest.Request) {
		handler.ServeHTTP(w.(http.ResponseWriter), r.Request)
	}))

	http.Handle("/git/", restAPI.MakeHandler())
}
