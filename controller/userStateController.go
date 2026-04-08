package controller

import (
	"errors"
	"net/http"
	apiresponse "ticketBookingSystem/internal/apiResponse"
	"ticketBookingSystem/repository"
	"ticketBookingSystem/service"

	"github.com/gin-gonic/gin"
)

type userStateController struct {
	service service.UserStateService
}

type UserStateController interface {
	GetState(ctx *gin.Context)
	UpdateState(ctx *gin.Context)
	ClearIntent(ctx *gin.Context)
}

// Constructor
func NewUserStateController(s service.UserStateService) *userStateController {
	return &userStateController{
		service: s,
	}
}

// GET /user-state/:userId
func (c *userStateController) GetState(ctx *gin.Context) {
	userID := ctx.GetString("user_id")
	if userID == "" {
		apiresponse.NewErrorResponse(ctx, errors.New("user id is required"), http.StatusBadRequest)
		return
	}

	state, err := c.service.GetState(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": state,
	})
}

// POST /user-state/:userId
func (c *userStateController) UpdateState(ctx *gin.Context) {
	userID := ctx.GetString("user_id")
	if userID == "" {
		apiresponse.NewErrorResponse(ctx, errors.New("user id is required"), http.StatusBadRequest)
		return
	}

	var state repository.UserState
	if err := ctx.ShouldBindJSON(&state); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	err := c.service.UpdateState(ctx, userID, state)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "state updated successfully",
	})
}

// DELETE /user-state/:userId/intent
func (c *userStateController) ClearIntent(ctx *gin.Context) {
	userID := ctx.GetString("user_id")
	if userID == "" {
		apiresponse.NewErrorResponse(ctx, errors.New("user id is required"), http.StatusBadRequest)
		return
	}

	err := c.service.ClearIntent(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "intent cleared",
	})
}
