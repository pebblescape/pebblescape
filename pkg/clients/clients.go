package clients

import (
	"fmt"
	"os"

	"github.com/fsouza/go-dockerclient"
)

func NewDockerClient() (*docker.Client, error) {
	endpoint := os.Getenv("DOCKER_HOST")
	path := os.Getenv("DOCKER_CERT_PATH")
	ca := fmt.Sprintf("%s/ca.pem", path)
	cert := fmt.Sprintf("%s/cert.pem", path)
	key := fmt.Sprintf("%s/key.pem", path)

	if path == "" {
		return docker.NewClient(endpoint)
	}

	return docker.NewTLSClient(endpoint, cert, key, ca)
}
