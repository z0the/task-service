package http

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"kroog-test/internal/task_service"
)

func NewController(lg *log.Logger, svc task_service.Service) *Controller {
	return &Controller{
		lg:  lg,
		svc: svc,
	}
}

type Controller struct {
	lg     *log.Logger
	svc    task_service.Service
	router *mux.Router
}

func (c *Controller) Run() error {
	c.initHandlers()

	return http.ListenAndServe(":3000", c.router)
}

func (c *Controller) initHandlers() {
	router := mux.NewRouter()

	router.Methods(http.MethodPost).HandlerFunc(c.scheduleTaskHandler)

	c.router = router
}

func (c *Controller) scheduleTaskHandler(w http.ResponseWriter, r *http.Request) {
	rawBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	task := &TaskDTO{}
	err = json.Unmarshal(rawBody, task)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errResp := &ErrorResp{
			Error: err.Error(),
		}
		rawErrResp, err := json.Marshal(errResp)
		if err != nil {
			c.lg.Println("marshal errResp error: ", err)
			return
		}
		_, err = w.Write(rawErrResp)
		if err != nil {
			c.lg.Println("write errResp error: ", err)
			return
		}
		return
	}

	err = c.svc.ScheduleTask(r.Context(), &task_service.Task{
		Time:    task.Time,
		QueueID: task.QueueID,
		Action:  task.Action,
	})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type ErrorResp struct {
	Error string `json:"error"`
}
