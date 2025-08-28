package controller

import "github.com/gin-gonic/gin"

type BookingController interface {
	GetBookings(ctx *gin.Context)
	ReserveMovieSeat(ctx *gin.Context)
	GetBookingStatus(ctx *gin.Context)
}

type bookingController struct{}

func (mc *bookingController) GetBookings(ctx *gin.Context) {

}
func (mc *bookingController) ReserveMovieSeat(ctx *gin.Context) {

}

func (mc *bookingController) GetBookingStatus(ctx *gin.Context) {

}

func NewBookingController() BookingController {
	return &bookingController{}
}
