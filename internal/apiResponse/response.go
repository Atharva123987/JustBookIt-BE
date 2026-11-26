package apiresponse

import (
	"errors"
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
	// Match with errors.Is instead of indexing the map with err: some error
	// types (e.g. mongo.CommandError) are not hashable and would panic.
	mappedErrorMessage := ""
	for domainErr, message := range constants.DomainErrorsMapped {
		if errors.Is(err, domainErr) {
			mappedErrorMessage = message
			break
		}
	}
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
