package main

import (
	"fmt"
	"niyodeploy/manager"
	"niyodeploy/node"
	"niyodeploy/task"
	"niyodeploy/worker"
	"os"
	"time"

	"github.com/docker/docker/client"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

func main() {
	task1 := task.Task{
		ID:     uuid.New(),
		Name:   "Task-1",
		State:  task.Pending,
		Image:  "nginx:latest",
		Memory: 512,
		Disk:   1,
	}

	tastEvent := task.TaskEvent{
		ID:        uuid.New(),
		State:     task.Pending,
		Timestamp: nil,
		Task:      &task1,
	}

	fmt.Printf("Task: %+v\n", task1)
	fmt.Printf("TaskEvent: %+v\n", tastEvent)

	w := worker.Worker{
		Name:  "Worker-1",
		Queue: *queue.New(),
		Db:    make(map[uuid.UUID]*task.Task),
	}

	fmt.Printf("Worker: %+v\n", w)
	w.CollectStats()
	w.RunTask()
	w.StartTask()
	w.StopTask()

	m := manager.Manager{
		Pending: *queue.New(),
		TaskDb:  make(map[string]*task.Task),
		EventDb: make(map[string]*task.TaskEvent),
		Workers: []string{w.Name},
	}

	fmt.Printf("Manager: %+v\n", m)
	m.SelectWorker()
	m.UpdateTasks()
	m.SendWork()

	n := node.Node{
		Name:   "Node-1",
		Ip:     "192.168.1.1",
		Cores:  4,
		Memory: 8192,
		Disk:   100,
		Role:   "worker",
	}

	fmt.Printf("Node: %+v\n", n)

	fmt.Printf("creating a test container...\n")
	docker, result := createContainer()
	if result.Error != nil {
		fmt.Printf("Error creating container: %v\n", result.Error)
		os.Exit(1)
		return
	}
	time.Sleep(10 * time.Second)
	_ = stopContainer(docker, result.ContainerID)
}

func createContainer() (*task.Docker, *task.DockerResult) {
	config := task.Config{
		Name:  "test-container-1",
		Image: "postgres:13",
		Env:   []string{"POSTGRES_USER=postgres", "POSTGRES_PASSWORD=postgres"},
	}

	dockerClient, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	docker := task.Docker{
		Client: dockerClient,
		Config: config,
	}

	result := docker.Run()
	if result.Error != nil {
		result.LogError()
		return nil, result
	}

	result.LogSuccess()

	return &docker, result
}

func stopContainer(docker *task.Docker, containerID string) *task.DockerResult {
	result := docker.Stop(containerID)
	if result.Error != nil {
		result.LogError()
		return result
	}
	result.LogSuccess()
	return result
}
