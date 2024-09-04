package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthController is a controller that handles health check requests.
//@Summary Health check endpoint
//@Description This endpoint is used to check the health of the server
//@Produce json
//@Success 200 {object} HealthController
//@Router /health [get]
//@Tags health

func (h Handler) Status(c *gin.Context) {
	Message := "Working!"
	c.JSON(http.StatusOK, Message)
}
