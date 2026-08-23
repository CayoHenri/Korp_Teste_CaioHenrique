package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type SuccessResponse struct {
	Success bool `json:"success" example:"true"`
	Data    any  `json:"data"`
}

type ErrorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Success bool      `json:"success" example:"false"`
	Error   ErrorData `json:"error"`
}

func Data(c *gin.Context, status int, data any) {
	c.JSON(status, SuccessResponse{
		Success: true,
		Data:    data,
	})
}

func OK(c *gin.Context, data any) {
	Data(c, http.StatusOK, data)
}

func Created(c *gin.Context, data any) {
	Data(c, http.StatusCreated, data)
}

func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, ErrorResponse{
		Success: false,
		Error: ErrorData{
			Code:    code,
			Message: message,
		},
	})
}
