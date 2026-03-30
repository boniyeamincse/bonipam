package http

import "github.com/gin-gonic/gin"

type Meta struct {
	RequestID string `json:"request_id"`
}

type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   interface{} `json:"error"`
	Meta    Meta        `json:"meta"`
}

func RespondOK(c *gin.Context, status int, data interface{}) {
	requestID := c.GetString("correlation_id")
	c.JSON(status, Envelope{
		Success: true,
		Data:    data,
		Error:   nil,
		Meta:    Meta{RequestID: requestID},
	})
}

func RespondError(c *gin.Context, status int, code, message string) {
	requestID := c.GetString("correlation_id")
	c.JSON(status, Envelope{
		Success: false,
		Data:    nil,
		Error: map[string]string{
			"code":    code,
			"message": message,
		},
		Meta: Meta{RequestID: requestID},
	})
}
