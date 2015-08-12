package host

type User struct {
	Name string   `json:"name,omitempty"`
	Keys []string `json:"keys,omitempty"`
}
