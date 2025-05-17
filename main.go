package main

import (
	"fmt"
	"niyodeploy/task"
	"niyodeploy/worker"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

func main() {
	db := make(map[uuid.UUID]*task.Task)
	w := worker.Worker{
		Queue: *queue.New(),
		Db:    db,
	}

	t := task.Task{
		ID:    uuid.New(),
		Name:  "test-container-1",
		State: task.Scheduled,
		Image: "strm/helloworld-http",
	}

	fmt.Println("starting task")
	w.AddTask(t)
	fmt.Println("task added to queue")
	result := w.RunTask()
	if result.Error != nil {
		panic(result.Error)
	}

	t.ContainerID = result.ContainerID
	time.Sleep(15 * time.Second)

	fmt.Println("stopping task")
	result = w.StopTask(t)
	if result.Error != nil {
		panic(result.Error)
	}
}
