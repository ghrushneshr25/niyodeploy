package task

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type Docker struct {
	Client *client.Client
	Config Config
}

type DockerResult struct {
	Error       error
	Action      string
	ContainerID string
	Result      string
}

func (d *Docker) Run() *DockerResult {
	log.Printf("Starting container for %s", d.Config.Name)
	ctx := context.Background()
	reader, err := d.Client.ImagePull(ctx, d.Config.Image, image.PullOptions{})
	if err != nil {
		return &DockerResult{Error: err, Action: "Pull Image"}
	}
	io.Copy(os.Stdout, reader)

	restartPolicy := container.RestartPolicy{
		Name: container.RestartPolicyMode(d.Config.RestartPolicy),
	}

	resources := container.Resources{
		Memory:   d.Config.Memory,
		NanoCPUs: int64(d.Config.Cpu * 1e9),
	}

	containerConfig := &container.Config{
		Image:        d.Config.Image,
		Tty:          false,
		Env:          d.Config.Env,
		ExposedPorts: d.Config.ExposedPorts,
	}

	hostConfig := container.HostConfig{
		RestartPolicy:   restartPolicy,
		Resources:       resources,
		PublishAllPorts: true,
	}

	resp, err := d.Client.ContainerCreate(ctx, containerConfig, &hostConfig, nil, nil, d.Config.Name)
	if err != nil {
		return &DockerResult{Error: err, Action: "Create Container"}
	}

	if err := d.Client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return &DockerResult{Error: err, Action: "Start Container"}
	}

	out, err := d.Client.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return &DockerResult{Error: err, Action: "Get Logs"}
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	return &DockerResult{
		ContainerID: resp.ID,
		Action:      "start",
		Result:      "success",
	}
}

func (d *Docker) Stop(containerID string) *DockerResult {
	log.Printf("Stopping container %s", containerID)
	ctx := context.Background()
	if err := d.Client.ContainerStop(ctx, containerID, container.StopOptions{}); err != nil {
		return &DockerResult{Error: err, Action: "Stop Container"}
	}

	if err := d.Client.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force:         false,
		RemoveVolumes: true,
		RemoveLinks:   false,
	}); err != nil {
		return &DockerResult{Error: err, Action: "Remove Container"}
	}

	return &DockerResult{
		ContainerID: containerID,
		Action:      "stop",
		Result:      "success",
		Error:       nil,
	}
}

func (dr *DockerResult) LogError() {
	log.Printf("Action: %s, ContainerID: %s, Error: %v", dr.Action, dr.ContainerID, dr.Error)
}

func (dr *DockerResult) LogSuccess() {
	log.Printf("Action: %s, ContainerID: %s, Result: %s", dr.Action, dr.ContainerID, dr.Result)
}
