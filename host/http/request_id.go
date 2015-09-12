package http

import (
	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/ant0ine/go-json-rest/rest"
	"github.com/pebblescape/pebblescape/pkg/random"
)

type RequestIdMiddleware struct {
}

func (mw *RequestIdMiddleware) MiddlewareFunc(h rest.HandlerFunc) rest.HandlerFunc {
	return func(w rest.ResponseWriter, r *rest.Request) {
		uuid := random.UUID()

		w.Header().Add("X-Request-ID", uuid)
		r.Env["REQUEST_ID"] = &uuid

		h(w, r)
	}
}
