package apiresponse

import (
	"ticketBookingSystem/internal/constants"

	"github.com/gin-gonic/gin"
)

type ApiResponse struct {
	Data       interface{} `json:"data"`
	Error      string      `json:"error,omitempty"`
	StatusCode int         `json:"-"`
}

func NewSuccessResponse(ctx *gin.Context, data interface{}, statusCode int) {
	resp := ApiResponse{
		Data:       data,
		StatusCode: statusCode,
	}
	ctx.JSON(statusCode, resp)
}

func NewErrorResponse(ctx *gin.Context, err error, statusCode int) {
	mappedErrorMessage := constants.DomainErrorsMapped[err]
	if mappedErrorMessage == "" {
		mappedErrorMessage = err.Error()
	}
	resp := ApiResponse{
		Data:       nil,
		Error:      mappedErrorMessage,
		StatusCode: statusCode,
	}
	ctx.JSON(statusCode, resp)
}
