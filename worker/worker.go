package worker

import (
	"errors"
	"fmt"
	"log"
	"niyodeploy/task"
	"time"

	"github.com/golang-collections/collections/queue"

	"github.com/google/uuid"
)

type Worker struct {
	Name      string
	Queue     queue.Queue
	Db        map[uuid.UUID]*task.Task
	TaskCount int
}

func (worker *Worker) CollectStats() {
	// Collect stats from the worker
	// This could include CPU usage, memory usage, etc.
	// For now, we'll just print a message
	println("Collecting stats from worker:", worker.Name)
}

func (worker *Worker) AddTask(t task.Task) {
	worker.Queue.Enqueue(t)
}

func (worker *Worker) RunTask() task.DockerResult {
	currentTask := worker.Queue.Dequeue()
	if currentTask == nil {
		log.Println("no tasks in the queue")
		return task.DockerResult{Error: errors.New("no tasks in queue"), Action: "RunTask"}
	}

	taskQueue := currentTask.(task.Task)
	taskPersisted := worker.Db[taskQueue.ID]
	if taskPersisted == nil { // If the task is not persisted, we need to persist it
		log.Printf("Persisting task: %v", taskQueue)
		taskPersisted = &taskQueue
		worker.Db[taskQueue.ID] = taskPersisted
	}
	var result task.DockerResult

	if task.ValidStateTransition(taskPersisted.State, task.Scheduled) {
		switch taskPersisted.State {
		case task.Scheduled:
			result = worker.StartTask(*taskPersisted)
		case task.Running:
			result = worker.StopTask(*taskPersisted)
		default:
			result.Error = errors.New("invalid state transition should not be possible")
		}
	} else {
		result.Error = fmt.Errorf("invalid transition from %v to %v", taskPersisted.State, task.Scheduled)
	}
	return result
}

func (worker *Worker) StartTask(t task.Task) task.DockerResult {
	t.StartTime = time.Now().UTC()
	config := t.NewConfig()
	docker, err := task.NewDocker(*config)
	if err != nil {
		log.Printf("failed to create Docker client: %v", err)
		t.State = task.Failed
		worker.Db[t.ID] = &t
		return task.DockerResult{Error: err, Action: "Create Docker Client"}
	}

	result := docker.Run()
	if result.Error != nil {
		result.LogError()
		t.State = task.Failed
		worker.Db[t.ID] = &t
		return result
	}

	t.ContainerID = result.ContainerID
	t.State = task.Running
	worker.Db[t.ID] = &t
	log.Printf("started ContainerID : %v task : %v", t.ContainerID, t.ID)
	return result
}

func (worker *Worker) StopTask(t task.Task) task.DockerResult {
	config := t.NewConfig()
	docker, err := task.NewDocker(*config)
	if err != nil {
		log.Printf("failed to create Docker client: %v", err)
		return task.DockerResult{Error: err, Action: "StopTask"}
	}

	result := docker.Stop(t.ContainerID)
	if result.Error != nil {
		result.LogError()
		return task.DockerResult{Error: result.Error, Action: "StopTask"}
	}

	t.FinishTime = time.Now().UTC()
	t.State = task.Completed
	worker.Db[t.ID] = &t
	log.Printf("Stopped ContainerID : %v task : %v", t.ContainerID, t.ID)
	return result
}

func (worker *Worker) GetTasks() []*task.Task {
	tasks := make([]*task.Task, 0)
	for _, task := range worker.Db {
		tasks = append(tasks, task)
	}
	return tasks
}
