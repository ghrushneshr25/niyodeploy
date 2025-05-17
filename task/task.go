package task

import (
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
)

type State int

const (
	Pending   State = iota // Initial state
	Scheduled              // A task moves to this state once the manager has scheduled it onto a worker.
	Running                // A task moves to this state once the worker has started it.
	Completed              // A task moves to this state once the worker has completed it (does not fail).
	Failed                 // A task moves to this state if it fails to start or run.
)

type Task struct {
	ID            uuid.UUID
	ContainerID   string
	Name          string
	State         State
	Image         string
	Memory        int
	Disk          int
	ExposedPorts  nat.PortSet
	PortBindings  map[string]string
	RestartPolicy string
	StartTime     time.Time
	FinishTime    time.Time
}

type TaskEvent struct {
	ID        uuid.UUID
	State     State
	Timestamp *time.Time
	Task      *Task
}

func (task *Task) NewConfig() *Config {
	return &Config{
		Name:          task.Name,
		Image:         task.Image,
		ExposedPorts:  task.ExposedPorts,
		RestartPolicy: task.RestartPolicy,
	}
}
