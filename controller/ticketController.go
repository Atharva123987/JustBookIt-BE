package controller

import "github.com/gin-gonic/gin"

type TicketController interface {
	GetTicket(ctx *gin.Context)
}

type ticketController struct{}

func (mc *ticketController) GetTicket(ctx *gin.Context) {

}

func NewTicketController() TicketController {
	return &ticketController{}
}
