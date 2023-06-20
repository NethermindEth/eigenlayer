package docker

//go:generate mockgen -package=mocks -destination=./mocks/apiClient.go github.com/docker/docker/client APIClient
