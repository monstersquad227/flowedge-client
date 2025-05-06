package service

import (
	"context"
	"encoding/json"
	"flowedge-client/utils"
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

func CreateContainers() (string, error) {
	dockerClient, err := newDockerClient()
	if err != nil {
		return "", err
	}
	result, err := dockerClient.ContainerCreate(context.Background(),
		&container.Config{
			User:  "root",
			Image: "harbor.chengduoduo.com/dev/auth:20250328_164355-2c271beb",
		},
		&container.HostConfig{
			Binds: []string{
				"/data/logs/:/data/logs",
				"/etc/localtime:/etc/localtime:ro",
			},
			NetworkMode: "host",
			Privileged:  true,
			RestartPolicy: container.RestartPolicy{
				Name:              "on-failure",
				MaximumRetryCount: 5,
			},
		}, nil, nil, utils.GetAgentID())
	if err != nil {
		return "", err
	}
	return result.ID, nil
}

func StartContainers() (string, error) {
	dockerClient, err := newDockerClient()
	if err != nil {
		return "", err
	}
	err = dockerClient.ContainerStart(context.Background(), utils.GetAgentID(), container.StartOptions{})
	if err != nil {
		return "", err
	}
	return utils.GetAgentID(), nil
}
