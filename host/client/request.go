package client

import (
	"encoding/json"
	"fmt"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/parnurzeal/gorequest"
	"github.com/pebblescape/pebblescape/pkg/version"
)

var pebblesHost = "http://localhost:4592"

type ServerError struct {
	Error string `json:"Error,omitempty"`
}

func SetHost(host string) {
	pebblesHost = host
}

func get(endpoint string) *gorequest.SuperAgent {
	return gorequest.New().Get(pebblesHost+endpoint).Set("User-Agent", "Pebblescape/"+version.String())
}

func post(endpoint string, content interface{}) *gorequest.SuperAgent {
	return gorequest.New().Post(pebblesHost+endpoint).Set("User-Agent", "Pebblescape/"+version.String()).Send(content)
}

func parseError(resp gorequest.Response, body []byte) error {
	if resp.StatusCode == 400 {
		var serverError ServerError
		err := json.Unmarshal(body, &serverError)
		if err != nil {
			return err
		}

		return fmt.Errorf("%s", serverError.Error)
	} else {
		return fmt.Errorf("Server error %v: %s", resp.StatusCode, body)
	}
}
