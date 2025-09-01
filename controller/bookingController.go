package controller

import (
	"errors"
	"net/http"
	apiresponse "ticketBookingSystem/internal/apiResponse"
	"ticketBookingSystem/service"

	"github.com/gin-gonic/gin"
)

type BookingController interface {
	GetBookings(ctx *gin.Context)
	ReserveMovieSeat(ctx *gin.Context)
	GetBookingStatus(ctx *gin.Context)
}

type bookingController struct {
	bookingService service.BookingService
}

func (mc *bookingController) GetBookings(ctx *gin.Context) {

}

type BookSeatRequest struct {
	ShowID      string   `json:"showId" binding:"required"`
	SeatNumbers []string `json:"seatNumbers" binding:"required"`
}

func (mc *bookingController) ReserveMovieSeat(ctx *gin.Context) {
	var req BookSeatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apiresponse.NewErrorResponse(ctx, err, http.StatusBadRequest)
		return
	}
	userID := ctx.GetString("user_id")
	if userID == "" {
		apiresponse.NewErrorResponse(ctx, errors.New("user id is required"), http.StatusBadRequest)
		return
	}
	res, err := mc.bookingService.ReserveMovieSeat(ctx, req.ShowID, req.SeatNumbers, userID)
	if err != nil {
		apiresponse.NewErrorResponse(ctx, err, http.StatusBadRequest)
		return
	}
	apiresponse.NewSuccessResponse(ctx, res, http.StatusOK)
}

func (mc *bookingController) GetBookingStatus(ctx *gin.Context) {

}

func NewBookingController(bookingService service.BookingService) BookingController {
	return &bookingController{bookingService}
}
