package repository

import (
	"context"

	"ticketBookingSystem/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MovieRepository interface {
	GetAllMovies(ctx context.Context, filter bson.M, opts ...*options.FindOptions) ([]models.Movie, error)
	GetMovieById(ctx context.Context, id string) (*models.Movie, error)
	GetLocationByCoordinates(ctx context.Context, location models.LocationMessage) ([]*models.Location, error)
	GetShowsByFilter(ctx context.Context, filter bson.M, opts ...*options.FindOptions) ([]models.Show, error)
}

type movieRepository struct {
	db *mongo.Database
}

func (mr *movieRepository) GetAllMovies(ctx context.Context, filter bson.M, opts ...*options.FindOptions) ([]models.Movie, error) {
	movies, err := mr.db.Collection("movie").Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}

	var result []models.Movie
	err = movies.All(ctx, &result)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return []models.Movie{}, nil
	}
	return result, nil
}

func (mr *movieRepository) GetMovieById(ctx context.Context, id string) (*models.Movie, error) {
	movie := &models.Movie{}
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	err = mr.db.Collection("movie").FindOne(ctx, bson.D{{"_id", objId}}).Decode(movie)
	if err != nil {
		return nil, err
	}
	return movie, nil
}

func (mr *movieRepository) GetLocationByCoordinates(ctx context.Context, location models.LocationMessage) ([]*models.Location, error) {
	filter := bson.M{
		"location": bson.M{
			"$near": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": []float64{location.Lng, location.Lat}, // user's location [lng, lat]
				},
				"$maxDistance": 10000, // in meters
			},
		},
	}
	var locations []*models.Location
	cursor, err := mr.db.Collection("location").Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &locations); err != nil {
		return nil, err
	}

	return locations, nil // return the nearest location if found
}

func (mr *movieRepository) GetShowsByFilter(ctx context.Context, filter bson.M, opts ...*options.FindOptions) ([]models.Show, error) {
	shows, err := mr.db.Collection("show").Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}

	var result []models.Show
	err = shows.All(ctx, &result)
	if err != nil {
		return nil, err
	}
	return result, nil // return all shows that match the filter criteria
}

func NewMovieRepository(client *mongo.Client) MovieRepository {
	return &movieRepository{db: client.Database("ticket-booking-system")}
}
