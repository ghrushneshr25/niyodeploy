package worker

import (
	"niyodeploy/task"

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

func (worker *Worker) RunTask() {
	// Run a task from the queue
	if worker.Queue.Len() > 0 {
		task := worker.Queue.Dequeue()
		if task != nil {
			println("Running task:", task)
			// Here you would run the task
			// For now, we'll just print a message
		} else {
			println("No tasks in the queue")
		}
	} else {
		println("No tasks in the queue")
	}
}

func (worker *Worker) StartTask() {
	// Start a task
	// This could involve pulling an image, starting a container, etc.
	// For now, we'll just print a message
	println("Starting task on worker:", worker.Name)
	worker.TaskCount++
}

func (worker *Worker) StopTask() {
	// Stop a task
	// This could involve stopping a container, removing an image, etc.
	// For now, we'll just print a message
	println("Stopping task on worker:", worker.Name)
	worker.TaskCount--
}
