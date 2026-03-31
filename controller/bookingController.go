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
	UpdateBookingStatus(ctx *gin.Context)
}

type bookingController struct {
	bookingService service.BookingService
}

func (mc *bookingController) GetBookings(ctx *gin.Context) {
	userID := ctx.GetString("user_id")
	if userID == "" {
		apiresponse.NewErrorResponse(ctx, errors.New("user id is required"), http.StatusBadRequest)
		return
	}
	bookings, err := mc.bookingService.GetBookings(ctx, userID)
	if err != nil {
		apiresponse.NewErrorResponse(ctx, err, http.StatusBadRequest)
		return
	}
	apiresponse.NewSuccessResponse(ctx, bookings, http.StatusOK)
}

type BookSeatRequest struct {
	ShowID      string   `json:"showId" binding:"required"`
	SeatNumbers []string `json:"seatNumbers"`
	NoOfSeats   int      `json:"noOfSeats" binding:"required"`
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
	res, err := mc.bookingService.ReserveMovieSeat(ctx, req.ShowID, req.NoOfSeats, req.SeatNumbers, userID)
	if err != nil {
		apiresponse.NewErrorResponse(ctx, err, http.StatusBadRequest)
		return
	}
	apiresponse.NewSuccessResponse(ctx, res, http.StatusOK)
}

func (mc *bookingController) UpdateBookingStatus(ctx *gin.Context) {

}

func NewBookingController(bookingService service.BookingService) BookingController {
	return &bookingController{bookingService}
}
