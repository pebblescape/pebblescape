package host

type User struct {
	Name  string `json:"name,omitempty"`
	Token string `json:"token,omitempty"`
}
