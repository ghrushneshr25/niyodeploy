package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"niyodeploy/task"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ApiRouter struct {
	Address string
	Port    int
	Worker  *Worker
	Router  *chi.Mux
}

func (apiRouter *ApiRouter) Start() {
	apiRouter.Init()
	log.Printf("starting api server on %s:%d", apiRouter.Address, apiRouter.Port)
	if err := http.ListenAndServe(fmt.Sprintf("%s:%d", apiRouter.Address, apiRouter.Port), apiRouter.Router); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func (apiRouter *ApiRouter) Init() {
	apiRouter.Router = chi.NewRouter()
	apiRouter.Router.Route("/tasks", func(r chi.Router) {
		r.Get("/", apiRouter.GetTaskHandler)
		r.Post("/", apiRouter.StartTaskHandler)
		r.Delete("/{id}", apiRouter.StopTaskHandler)
	})
}

func (apiRouter *ApiRouter) StartTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	var taskEvent task.TaskEvent
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&taskEvent); err != nil {
		log.Printf("error unmarshalling body: %v", err)
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusBadRequest)
		return
	}

	if taskEvent.Task == nil {
		http.Error(w, `{"error":"missing task in request body"}`, http.StatusBadRequest)
		return
	}

	taskEvent.Task.ID = uuid.New()

	apiRouter.Worker.AddTask(*taskEvent.Task)
	log.Printf("task added to queue: %v", taskEvent.Task)
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(taskEvent.Task); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}
func (apiRouter *ApiRouter) StopTaskHandler(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	w.Header().Set("Content-Type", "application/json")

	if taskID == "" {
		http.Error(w, `{"error":"missing task id in request"}`, http.StatusBadRequest)
		return
	}

	taskIDParsed, err := uuid.Parse(taskID)
	if err != nil {
		http.Error(w, `{"error":"invalid task id"}`, http.StatusBadRequest)
		return
	}

	taskToStop, ok := apiRouter.Worker.Db[taskIDParsed]
	if !ok {
		http.Error(w, `{"error":"task not found"}`, http.StatusNotFound)
		return
	}

	taskCopy := *taskToStop
	taskCopy.State = task.Completed
	apiRouter.Worker.Db[taskIDParsed] = &taskCopy
	apiRouter.Worker.AddTask(taskCopy)
	log.Printf("added task %v to stop container %v\n", taskIDParsed, taskCopy.ContainerID)
	w.WriteHeader(204)
}
func (apiRouter *ApiRouter) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiRouter.Worker.GetTasks())
}
