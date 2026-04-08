package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"ticketBookingSystem/models"
	"ticketBookingSystem/tools"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type BookingRepository interface {
	GetSeats(ctx context.Context, showId string, noOfSeats int, seatNumbers []string) ([]*models.Seat, []string, error)
	UpdateMultipleSeatsStatus(
		ctx context.Context,
		showId string,
		seatNumbers []string,
		newStatus string,
		bookingId string,
	) error
	CreateBooking(
		ctx context.Context,
		booking *models.Booking,
	) (primitive.ObjectID, error)
	GetBookingByUserId(ctx context.Context, userId string) ([]*models.Booking, error)
	GetBookingByBookingId(ctx context.Context, bookingId string) (*models.Booking, error)
	UpdateBooking(ctx context.Context, update bson.M, filter bson.M) error
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

	// Lookup map
	lookup := make(map[string]bool, len(seatNumbers))
	for _, num := range seatNumbers {
		lookup[num] = true
	}

	for i := range rows {
		row := &rows[i]
		seatIndex := 1 // numbering inside row (skip AISLE)

		for j := range row.Seats {
			seat := &row.Seats[j]

			// Skip AISLE completely
			if seat.Type == "AISLE" {
				continue
			}

			seatNum := seat.SeatNumber

			// 🔥 Generate if missing
			if seatNum == "" {
				seatNum = row.Row + strconv.Itoa(seatIndex)
			}

			if lookup[seatNum] {
				result = append(result, seat)
			}

			seatIndex++
		}
	}

	return result
}

func (mr *bookingRepository) GetSeats(ctx context.Context, showId string, noOfSeats int, seatNumbers []string) ([]*models.Seat, []string, error) {
	show, err := mr.movieRepo.GetShowById(ctx, showId)
	if err != nil {
		return nil, seatNumbers, err
	}
	if noOfSeats != 0 && len(seatNumbers) == 0 {
		// Select seats according to algorithm
		seatNumbers, err = tools.FindBestSeats(show, noOfSeats)
		if err != nil {
			return nil, seatNumbers, err
		}
	}
	seats := findSeats(show.SeatRows, seatNumbers)
	if seats == nil {
		return nil, seatNumbers, errors.New("seat not found")
	}

	return seats, seatNumbers, nil // return all shows that match the filter criteria
}

func seatNumberToIndex(seatNumber string) (int, int, error) {
	if len(seatNumber) < 2 {
		return -1, -1, fmt.Errorf("invalid seat number")
	}

	// Row index
	rowIndex := int(seatNumber[0] - 'A')

	// Column index
	colNum, err := strconv.Atoi(seatNumber[1:])
	if err != nil {
		return -1, -1, err
	}
	colIndex := colNum - 1

	return rowIndex, colIndex, nil
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

	var updates []mongo.WriteModel

	for _, seatNum := range seatNumbers {

		rowIndex, colIndex, err := seatNumberToIndex(seatNum)
		if err != nil {
			return err
		}

		filter := bson.M{
			"_id": objID,
			fmt.Sprintf("seat_rows.%d.seats.%d.status", rowIndex, colIndex): "available",
		}

		update := bson.M{
			"$set": bson.M{
				fmt.Sprintf("seat_rows.%d.seats.%d.status", rowIndex, colIndex):     newStatus,
				fmt.Sprintf("seat_rows.%d.seats.%d.booking_id", rowIndex, colIndex): bookingId,
			},
		}

		model := mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update)

		updates = append(updates, model)
	}

	res, err := mr.db.Collection("show").BulkWrite(ctx, updates)
	if err != nil {
		return err
	}

	if res.ModifiedCount != int64(len(seatNumbers)) {
		return errors.New("some seats are already booked")
	}

	return nil
}
func (mr *bookingRepository) CreateBooking(
	ctx context.Context,
	booking *models.Booking,
) (primitive.ObjectID, error) {

	res, err := mr.db.Collection("booking").InsertOne(ctx, booking)
	if err != nil {
		return primitive.NilObjectID, err
	}

	// ✅ Type assertion
	id, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID, fmt.Errorf("failed to convert inserted ID to ObjectID")
	}

	return id, nil
}

func (mr *bookingRepository) GetBookingByUserId(ctx context.Context, userId string) ([]*models.Booking, error) {
	filter := bson.M{"user_id": userId}
	bookings, err := mr.db.Collection("booking").Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var result []*models.Booking
	err = bookings.All(ctx, &result)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return []*models.Booking{}, nil
	}
	return result, nil
}
func (mr *bookingRepository) GetBookingByBookingId(ctx context.Context, bookingId string) (*models.Booking, error) {
	filter := bson.M{"booking_id": bookingId}
	booking, err := mr.db.Collection("booking").Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var result []*models.Booking
	err = booking.All(ctx, &result)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}
	return result[0], nil
}
func (mr *bookingRepository) UpdateBooking(
	ctx context.Context,
	update bson.M,
	filter bson.M,
) error {
	updateM := bson.M{
		"$set": update,
	}
	res, err := mr.db.Collection("booking").UpdateOne(ctx, filter, updateM)
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return errors.New("booking not found")
	}

	return nil
}
