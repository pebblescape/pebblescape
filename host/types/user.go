package host

import (
	"github.com/pebblescape/pebblescape/pkg/random"
)

type User struct {
	Name  string `json:"name,omitempty"`
	Token string `json:"token,omitempty"`
}

func (u *User) Create() error {
	u.Token = random.Hex(20)
	return nil
}
