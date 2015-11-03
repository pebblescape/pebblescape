package http

import (
	"log"
	"net"
	"net/http"
	"path/filepath"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/ant0ine/go-json-rest/rest"
	hostApi "github.com/pebblescape/pebblescape/host/api"
	"github.com/pebblescape/pebblescape/host/config"
	"github.com/pebblescape/pebblescape/host/gitreceive"
	"github.com/pebblescape/pebblescape/pkg/shutdown"
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

func Serve(port string, hostApi *hostApi.Api, conf *config.Config, logger *log.Logger) error {
	DefaultStack = append(
		[]rest.Middleware{
			&rest.AccessLogApacheMiddleware{
				Logger: logger,
				Format: LogFormat,
			},
		}, DefaultStack...)

	auth := func(user string, password string) bool {
		return password == conf.HostKey
	}

	setupGit(auth, conf)
	setupHttp(auth, hostApi)

	// Create server
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	shutdown.BeforeExit(func() { listener.Close() })

	log.Println("HTTP API listening on " + port)
	return http.Serve(listener, nil)
}

func setupHttp(auth func(string, string) bool, hostApi *hostApi.Api) {
	api := rest.NewApi()
	handler := Handler{hostApi}

	router, err := rest.MakeRouter(
		rest.Get("/job", handler.ListJobs),
	)
	if err != nil {
		log.Fatal(err)
	}

	api.Use(DefaultStack...)
	api.Use(&rest.AuthBasicMiddleware{
		Realm:         "Pebblescape",
		Authenticator: auth,
	})
	api.Use(&rest.ContentTypeCheckerMiddleware{})
	api.SetApp(router)

	http.Handle("/", api.MakeHandler())
}

func setupGit(auth func(string, string) bool, conf *config.Config) {
	cachePath := filepath.Join(conf.Home, "cache")
	repoPath := filepath.Join(conf.Home, "repos")

	api := rest.NewApi()
	handler := gitreceive.NewGitHandler(repoPath, cachePath)

	api.Use(DefaultStack...)
	api.Use(&rest.AuthBasicMiddleware{
		Realm:         "Pebblescape GIT",
		Authenticator: auth,
	})
	api.SetApp(rest.AppSimple(func(w rest.ResponseWriter, r *rest.Request) {
		handler.ServeHTTP(w.(http.ResponseWriter), r.Request)
	}))

	http.Handle("/git/", api.MakeHandler())
}
