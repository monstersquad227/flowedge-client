package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flowedge-client/utils"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	image2 "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"io/ioutil"
	"strings"
)

func newDockerClient() (*client.Client, error) {
	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return apiClient, nil
}

func containerDragon(image, agentID string) (string, error) {
	dockerClient, err := newDockerClient()
	if err != nil {
		return "", err
	}
	_, err = PullImage(image)
	if err != nil {
		return "", err
	}

	args := filters.NewArgs()
	args.Add("name", "/"+agentID)

	containers, err := dockerClient.ContainerList(context.Background(), container.ListOptions{All: true, Filters: args})
	if err != nil {
		return "", err
	}
	for _, value := range containers {
		err = dockerClient.ContainerRemove(context.Background(), value.ID, container.RemoveOptions{Force: true})
		if err != nil {
			return "", err
		}
	}

	_, err = CreateContainers(image)
	if err != nil {
		return "", err
	}

	result, err := StartContainers()
	if err != nil {
		return "", err
	}
	return result, nil
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

func CreateContainers(image string) (string, error) {
	dockerClient, err := newDockerClient()
	if err != nil {
		return "", err
	}
	result, err := dockerClient.ContainerCreate(context.Background(),
		&container.Config{
			User:  "root",
			Image: image,
			//ExposedPorts: nat.PortSet{
			//	"8500/tcp": struct{}{},
			//},
			Env: []string{
				"JVM_OPTIONS=" + utils.JVMOPTIONS,
				"APM_OPTIONS=" + utils.APMOPTIONS,
				"SERVER_IP=" + utils.GetAddress(),
			},
		},
		&container.HostConfig{
			// Linux 服务器映射目录
			Binds: []string{
				"/data/logs/:/data/logs",
				"/etc/localtime:/etc/localtime:ro",
			},
			// host 模式
			NetworkMode: "host",
			Privileged:  true,
			RestartPolicy: container.RestartPolicy{
				Name:              "on-failure",
				MaximumRetryCount: 5,
			},
			// 不使用host模式的端口映射
			//PortBindings: nat.PortMap{
			//	"8500/tcp": []nat.PortBinding{
			//		{
			//			HostIP:   "0.0.0.0",
			//			HostPort: "8500",
			//		},
			//	},
			//},
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

func StopContainer(containerID string) (string, error) {
	dockerClient, err := newDockerClient()
	if err != nil {
		return "", err
	}
	err = dockerClient.ContainerStop(context.Background(), containerID, container.StopOptions{})
	if err != nil {
		return "", err
	}
	return "Ok", nil
}

func RemoveContainer(containerID string) (string, error) {
	dockerClient, err := newDockerClient()
	if err != nil {
		return "", err
	}
	err = dockerClient.ContainerRemove(context.Background(), containerID, container.RemoveOptions{Force: true})
	if err != nil {
		return "", err
	}
	return "Ok", nil
}

func PullImage(image string) (string, error) {
	dockerClient, err := newDockerClient()
	if err != nil {
		return "", err
	}

	type MyAuthConfig struct {
		Username      string
		Password      string
		ServerAddress string
	}
	imageSplit := strings.Split(image, "")
	var auth MyAuthConfig
	auth.Username = "admin"
	auth.Password = "cdd.1q2w3e4r"
	auth.ServerAddress = "https://" + imageSplit[0]
	authMap := map[string]string{
		"username":      auth.Username,
		"password":      auth.Password,
		"serverAddress": auth.ServerAddress,
	}
	authJSON, err := json.Marshal(authMap)
	if err != nil {
		return "", err
	}
	result, err := dockerClient.ImagePull(context.Background(), image, image2.PullOptions{RegistryAuth: base64.StdEncoding.EncodeToString(authJSON)})
	if err != nil {
		return "", err
	}
	defer result.Close()
	bytes, err := ioutil.ReadAll(result)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func ListImage() (string, error) {
	dockerClient, err := newDockerClient()
	if err != nil {
		return "", err
	}
	images, err := dockerClient.ImageList(context.Background(), image2.ListOptions{})
	if err != nil {
		return "", err
	}
	bytes, err := json.MarshalIndent(images, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
