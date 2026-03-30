package tracing

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const HeaderCorrelationID = "X-Correlation-ID"

func CorrelationIDFromRequest(c *gin.Context) string {
	id := c.GetHeader(HeaderCorrelationID)
	if id == "" {
		id = uuid.NewString()
	}
	return id
}
