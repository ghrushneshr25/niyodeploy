package main

import (
	"log"
	"niyodeploy/task"
	"niyodeploy/worker"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

func main() {
	host := "localhost"
	port := 8080

	w := worker.Worker{
		Queue: *queue.New(),
		Db:    make(map[uuid.UUID]*task.Task),
	}

	api := worker.ApiRouter{
		Address: host,
		Port:    port,
		Worker:  &w,
	}

	go runTasks(&w)
	go w.CollectStats()
	api.Start()
}

func runTasks(worker *worker.Worker) {
	for {
		if worker.Queue.Len() != 0 {
			result := worker.RunTask()
			if result.Error != nil {
				log.Println("Error running task:", result.Error)
			}
		} else {
			log.Printf("no tasks to process currently.\n")
		}
		log.Println("sleeping for 5 seconds...")
		time.Sleep(5 * time.Second)
	}
}
