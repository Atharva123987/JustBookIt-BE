package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ---- Model ----

type UserState struct {
	// UserID      string    `bson:"user_id" json:"user_id"`
	DeviceID    string    `bson:"device_id" json:"device_id"`
	Movie       *string   `bson:"movie,omitempty" json:"movie,omitempty"`
	NoOfSeats   *int      `bson:"no_of_seats,omitempty" json:"noOfSeats,omitempty"`
	SeatNumbers []string  `bson:"seat_numbers,omitempty" json:"seatNumbers,omitempty"`
	MovieID     *string   `bson:"movie_id,omitempty" json:"movieId,omitempty"`
	Intent      *string   `bson:"intent,omitempty" json:"intent,omitempty"`
	ShowID      *string   `bson:"show_id,omitempty" json:"showId,omitempty"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`
}

// ---- Interface ----

type UserStateRepository interface {
	GetState(ctx context.Context, userID string) (*UserState, error)
	UpdateState(ctx context.Context, userID string, state UserState) error
	ClearIntent(ctx context.Context, userID string) error
}

// ---- Implementation ----

type userStateRepository struct {
	db *mongo.Database
}

// Constructor (same pattern as MovieRepository)
func NewUserStateRepository(client *mongo.Client) UserStateRepository {
	return &userStateRepository{
		db: client.Database("ticket-booking-system"),
	}
}

// ---- Methods ----

// Get state by device_id
func (r *userStateRepository) GetState(ctx context.Context, deviceId string) (*UserState, error) {
	var state UserState

	err := r.db.Collection("user_state").
		FindOne(ctx, bson.M{"device_id": deviceId}).
		Decode(&state)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("state not found")
		}
		return nil, err
	}

	return &state, nil
}

// Upsert state
func (r *userStateRepository) UpdateState(ctx context.Context, deviceId string, state UserState) error {
	state.DeviceID = deviceId
	state.UpdatedAt = time.Now()

	update := bson.M{
		"$set": state,
	}

	opts := options.Update().SetUpsert(true)

	_, err := r.db.Collection("user_state").UpdateOne(
		ctx,
		bson.M{"device_id": deviceId},
		update,
		opts,
	)

	return err
}

// Clear only intent (and keep rest of state)
func (r *userStateRepository) ClearIntent(ctx context.Context, deviceId string) error {
	update := bson.M{
		"$unset": bson.M{
			"intent": "",
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	result, err := r.db.Collection("user_state").UpdateOne(
		ctx,
		bson.M{"device_id": deviceId},
		update,
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("state not found")
	}

	return nil
}
