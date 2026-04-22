package service

import (
	"context"
	"crypto/rand"
	"math/big"
	"ticketBookingSystem/internal/constants"
	"ticketBookingSystem/models"
	"ticketBookingSystem/repository"

	"go.mongodb.org/mongo-driver/bson"
)

type BookingService interface {
	ReserveMovieSeat(ctx context.Context, showId string, noOfSeats int, seatNumbers []string, userId string) (*models.InitiatePaymentResponse, error)
	GetBookings(ctx context.Context, userId string) ([]*models.Booking, error)
	ConfirmBooking(ctx context.Context, userId string, bookingId string) error
	GetBookingByBookingId(ctx context.Context, bookingId string) (*models.Booking, error)
}

type bookingService struct {
	bookingRepository repository.BookingRepository
	paymentService    PaymentService
	movieRepository   repository.MovieRepository
}

func NewBookingService(bookingRepository repository.BookingRepository, paymentService PaymentService, movieRepository repository.MovieRepository) BookingService {
	return &bookingService{bookingRepository, paymentService, movieRepository}
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

	show, err := bs.movieRepository.GetShowById(ctx, showId)
	if err != nil {
		return nil, err
	}

	movieId := show.MovieID

	movie, err := bs.movieRepository.GetMovieById(ctx, movieId)
	if err != nil {
		return nil, err
	}

	booking := &models.Booking{
		ShowId:          showId,
		SeatNumbers:     seatNumbers,
		BookingId:       bookingId,
		UserId:          userId,
		MovieName:       movie.Title,
		ShowTimingEpoch: show.ShowTimingEpoch,
		MoviePoster:     movie.PosterURL,
		Status:          constants.BookingStatusPending,
		PaymentInfo: models.PaymentInfo{
			PaymentStatus: constants.PaymentStatusPending,
			PaymentMethod: constants.Stripe,
		},
	}
	booking.PaymentInfo.PaymentAmount = float64(len((seatNumbers))) * 100

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

	res, err := bs.paymentService.InitiatePayment(booking.PaymentInfo.PaymentAmount, userId, bookingId)
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
func (bs *bookingService) GetBookingByBookingId(ctx context.Context, bookingId string) (*models.Booking, error) {
	booking, err := bs.bookingRepository.GetBookingByBookingId(ctx, bookingId)
	if err != nil {
		return nil, err
	}
	return booking, nil
}

func (bs *bookingService) ConfirmBooking(ctx context.Context, userId string, bookingId string) error {
	filter := bson.M{"booking_id": bookingId}
	update := bson.M{
		"status":                      "confirmed",
		"payment_info.payment_status": "success",
	}
	err := bs.bookingRepository.UpdateBooking(ctx, update, filter)
	return err
}
