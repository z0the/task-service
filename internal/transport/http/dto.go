package http

type TaskDTO struct {
	Time    int64  `json:"time"`
	QueueID string `json:"queue_id"`
	Action  string `json:"action"`
}
