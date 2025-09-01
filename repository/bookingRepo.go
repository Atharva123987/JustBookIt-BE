package repository

import (
	"context"
	"errors"
	"ticketBookingSystem/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BookingRepository interface {
	GetSeats(ctx context.Context, showId string, seatNumbers []string) ([]*models.Seat, error)
	UpdateMultipleSeatsStatus(
		ctx context.Context,
		showId string,
		seatNumbers []string,
		newStatus string,
		bookingId string,
	) error
	CreateBooking(ctx context.Context, booking *models.Booking) error
}

type bookingRepository struct {
	db        *mongo.Database
	movieRepo MovieRepository
}

func NewBookingRepository(client *mongo.Client, movieRepo MovieRepository) BookingRepository {
	return &bookingRepository{db: client.Database("ticket-booking-system"), movieRepo: movieRepo}
}

func findSeats(rows []models.SeatRow, seatNumbers []string) []*models.Seat {
	var result []*models.Seat

	// Create a lookup map for faster seat number checks
	lookup := make(map[string]bool, len(seatNumbers))
	for _, num := range seatNumbers {
		lookup[num] = true
	}

	for i := range rows {
		for j := range rows[i].Seats {
			if lookup[rows[i].Seats[j].SeatNumber] {
				result = append(result, &rows[i].Seats[j])
			}
		}
	}

	return result
}

func (mr *bookingRepository) GetSeats(ctx context.Context, showId string, seatNumbers []string) ([]*models.Seat, error) {

	show, err := mr.movieRepo.GetShowById(ctx, showId)
	if err != nil {
		return nil, err
	}

	seats := findSeats(show.SeatRows, seatNumbers)
	if seats == nil {
		return nil, errors.New("seat not found")
	}

	return seats, nil // return all shows that match the filter criteria
}

func (mr *bookingRepository) UpdateMultipleSeatsStatus(
	ctx context.Context,
	showId string,
	seatNumbers []string,
	newStatus string,
	bookingId string,
) error {
	objID, err := primitive.ObjectIDFromHex(showId)
	if err != nil {
		return err
	}

	// ✅ Ensure we only update seats that are "available"
	filter := bson.M{
		"_id": objID,
		"seat_rows.seats": bson.M{
			"$elemMatch": bson.M{
				"seat_number": bson.M{"$in": seatNumbers},
				"status":      "available", // only update if still available
			},
		},
	}

	update := bson.M{
		"$set": bson.M{
			"seat_rows.$[].seats.$[seat].status":     newStatus,
			"seat_rows.$[].seats.$[seat].booking_id": bookingId,
		},
	}

	opts := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"seat.seat_number": bson.M{"$in": seatNumbers}, "seat.status": "available"},
		},
	})

	res, err := mr.db.Collection("show").UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}

	// If ModifiedCount == 0, none of the requested seats were available
	if res.ModifiedCount == 0 {
		return errors.New("some or all seats are already reserved")
	}

	return nil
}

func (mr *bookingRepository) CreateBooking(ctx context.Context, booking *models.Booking) error {
	_, err := mr.db.Collection("booking").InsertOne(ctx, booking)
	if err != nil {
		return err
	}

	return nil
}
