package service

import (
	"context"
	"encoding/json"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func newDockerClient() (*client.Client, error) {
	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return apiClient, nil
}

func ListContainers() (string, error) {
	dockerClient, err := newDockerClient()
	if err != nil {
		return "", err
	}
	containerList, err := dockerClient.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return "", err
	}
	bytes, err := json.MarshalIndent(containerList, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
