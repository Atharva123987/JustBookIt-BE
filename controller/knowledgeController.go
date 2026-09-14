package controller

import (
	"errors"
	"net/http"
	apiresponse "ticketBookingSystem/internal/apiResponse"
	"ticketBookingSystem/models"
	"ticketBookingSystem/service"

	"github.com/gin-gonic/gin"
)

type KnowledgeController interface {
	CreateChunk(ctx *gin.Context)
}

type knowledgeController struct {
	knowledgeService service.KnowledgeService
}

func (mc *knowledgeController) CreateChunk(ctx *gin.Context) {
	var body models.KnowledgeChunk
	if err := ctx.ShouldBindJSON(&body); err != nil {
		apiresponse.NewErrorResponse(ctx, err, http.StatusBadRequest)
		return
	}
	if body.Content == "" {
		apiresponse.NewErrorResponse(ctx, errors.New("Please verify data"), http.StatusBadRequest)
		return
	}
	createdChunk, err := mc.knowledgeService.CreateKnowledgeChunk(ctx, body)
	if err != nil {
		apiresponse.NewErrorResponse(ctx, err, http.StatusBadRequest)
		return
	}
	apiresponse.NewSuccessResponse(ctx, createdChunk, http.StatusOK)
}

func NewKnowledgeController(knowledgeService service.KnowledgeService) KnowledgeController {
	return &knowledgeController{
		knowledgeService: knowledgeService,
	}
}
