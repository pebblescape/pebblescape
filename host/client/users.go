package client

import (
	"encoding/json"
	"fmt"

	"github.com/pebblescape/pebblescape/host/types"
)

func ListUsers() ([]host.User, error) {
	resp, body, errs := get("/user").EndBytes()
	if errs != nil {
		return nil, errs[0]
	}
	if resp.StatusCode != 200 {
		return nil, parseError(resp, body)
	}

	var users []host.User
	err := json.Unmarshal(body, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func GetUser(username string) (*host.User, error) {
	resp, body, errs := get("/user/" + username).EndBytes()
	if errs != nil {
		return nil, errs[0]
	}
	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("User not found")
	}
	if resp.StatusCode != 200 {
		return nil, parseError(resp, body)
	}

	var user host.User
	err := json.Unmarshal(body, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func CreateUser(user host.User) error {
	resp, body, errs := post("/user", user).EndBytes()
	if errs != nil {
		return errs[0]
	}
	if resp.StatusCode != 200 {
		return parseError(resp, body)
	}

	return nil
}
