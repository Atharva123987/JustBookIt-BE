package service

import (
	"context"
	"crypto/rand"
	"math/big"
	"ticketBookingSystem/internal/constants"
	"ticketBookingSystem/models"
	"ticketBookingSystem/repository"
)

type BookingService interface {
	ReserveMovieSeat(ctx context.Context, showId string, noOfSeats int, seatNumbers []string, userId string) (*models.InitiatePaymentResponse, error)
	GetBookings(ctx context.Context, userId string) ([]*models.Booking, error)
}

type bookingService struct {
	bookingRepository repository.BookingRepository
	paymentService    PaymentService
}

func NewBookingService(bookingRepository repository.BookingRepository, paymentService PaymentService) BookingService {
	return &bookingService{bookingRepository, paymentService}
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateBookingId() string {
	result := make([]byte, 12)
	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		result[i] = letters[num.Int64()]
	}
	return string(result)
}

func (bs *bookingService) reserveSeat(ctx context.Context, showId string, seatNumbers []string, bookingId string) error {

	err := bs.bookingRepository.UpdateMultipleSeatsStatus(ctx, showId, seatNumbers, constants.SeatStatusReserved, bookingId)
	if err != nil {
		return err
	}
	return nil
}

func (bs *bookingService) createBookingRecord(ctx context.Context, showId string, seatNumbers []string, bookingId string, userId string) (*models.Booking, error) {
	var paymentAmount float64 = 0.0

	booking := &models.Booking{
		ShowId:      showId,
		SeatNumbers: seatNumbers,
		BookingId:   bookingId,
		UserId:      userId,
		Status:      constants.BookingStatusPending,
		PaymentInfo: models.PaymentInfo{
			PaymentStatus: constants.PaymentStatusPending,
			PaymentAmount: paymentAmount,
		},
	}

	id, err := bs.bookingRepository.CreateBooking(ctx, booking)
	if err != nil {
		return nil, err
	}
	booking.ID = id
	return booking, nil
}

func (bs *bookingService) ReserveMovieSeat(ctx context.Context, showId string, noOfSeats int, seatNumbers []string, userId string) (*models.InitiatePaymentResponse, error) {
	bookingId := generateBookingId()

	seats, seatNumbers, err := bs.bookingRepository.GetSeats(ctx, showId, noOfSeats, seatNumbers)
	if err != nil {
		return nil, err
	}
	if len(seats) == 0 {
		return nil, constants.ErrSeatNotFound
	}
	if len(seats) != noOfSeats {
		return nil, constants.ErrInvalidSeatNum
	}
	for _, seat := range seats {
		if seat.Status == constants.SeatStatusReserved {
			return nil, constants.ErrSeatReserved
		}
	}

	err = bs.reserveSeat(ctx, showId, seatNumbers, bookingId)
	if err != nil {
		return nil, err
	}
	// create booking record
	booking, err := bs.createBookingRecord(ctx, showId, seatNumbers, bookingId, userId)
	if err != nil {
		return nil, err
	}

	res, err := bs.paymentService.InitiatePayment(booking.PaymentInfo.PaymentAmount, userId)
	if err != nil {
		return nil, err
	}
	res.Booking = booking
	res.SeatNumbers = seatNumbers
	return res, nil
}

func (bs *bookingService) GetBookings(ctx context.Context, userId string) ([]*models.Booking, error) {
	bookings, err := bs.bookingRepository.GetBookingByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}
	return bookings, nil
}
