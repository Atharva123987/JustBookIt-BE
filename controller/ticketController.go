package controller

import "github.com/gin-gonic/gin"

type TicketController interface {
	GetTicket(ctx *gin.Context)
	VerifyTicket(ctx *gin.Context)
}

type ticketController struct{}

func (mc *ticketController) GetTicket(ctx *gin.Context) {

}

func (mc *ticketController) VerifyTicket(ctx *gin.Context) {

}

func NewTicketController() TicketController {
	return &ticketController{}
}
