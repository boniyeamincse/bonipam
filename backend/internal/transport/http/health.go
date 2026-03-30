package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HealthHandler(serviceName, env string) gin.HandlerFunc {
	return func(c *gin.Context) {
		RespondOK(c, http.StatusOK, gin.H{
			"service": serviceName,
			"status":  "ok",
			"env":     env,
		})
	}
}
