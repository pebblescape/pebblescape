package host

type App struct {
	Name      string `json:"name,omitempty"`
	OwnerName string `json:"user,omitempty"`
}
