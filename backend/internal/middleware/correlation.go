package middleware

import (
	"boni-pam/pkg/tracing"

	"github.com/gin-gonic/gin"
)

func CorrelationID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := tracing.CorrelationIDFromRequest(c)
		c.Writer.Header().Set(tracing.HeaderCorrelationID, id)
		c.Set("correlation_id", id)
		c.Next()
	}
}
