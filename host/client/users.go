package client

import (
	"encoding/json"
	"fmt"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/parnurzeal/gorequest"
	"github.com/pebblescape/pebblescape/host/types"
)

func ListUsers() ([]host.User, error) {
	resp, body, errs := gorequest.New().Get("http://localhost:4592/user").EndBytes()
	if errs != nil {
		return nil, errs[0]
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("non 200 response code: %v, %s", resp.StatusCode, body)
	}

	var users []host.User
	err := json.Unmarshal(body, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}
