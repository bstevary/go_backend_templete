package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthResponse struct {
	WorkerID int64  `json:"worker_id"`
	State    string `json:"state"`
}

func (h Handler) Status(c *gin.Context) {
	resp := HealthResponse{
		WorkerID: h.config.WorkerID,
		State:    "OK",
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Service is up and running",
		Data:    resp,
	})
}
