package service

import (
	"errors"
	"ticketBookingSystem/models"
	"ticketBookingSystem/repository"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MovieService interface {
	GetAllMovies(ctx *gin.Context, limit int, lastId string, searchText string) ([]models.Movie, error)
	GetMovieDetails(ctx *gin.Context, movieId string) (*models.Movie, error)
	GetShowsByLocation(ctx *gin.Context, location models.LocationMessage, movieId string) ([]models.LocationWithShow, error)
}

type movieService struct {
	movieRepo repository.MovieRepository
}

func NewMovieService(repo repository.MovieRepository) MovieService {
	return &movieService{movieRepo: repo}
}

func (ms *movieService) GetMovieDetails(ctx *gin.Context, movieId string) (*models.Movie, error) {
	return ms.movieRepo.GetMovieById(ctx, movieId)
}

func (ms *movieService) GetAllMovies(ctx *gin.Context, limit int, lastId string, searchText string) ([]models.Movie, error) {
	if limit < 1 || limit > 100 {
		limit = 10
	}
	filter := bson.M{}
	proj := bson.M{"_id": 1, "title": 1, "release_date": 1, "poster_url": 1}
	opts := options.Find().SetLimit(int64(limit)).SetProjection(proj)

	if lastId != "" {
		id, err := primitive.ObjectIDFromHex(lastId)
		if err != nil {
			return nil, err
		}
		filter["_id"] = bson.M{"$gt": id}
	}
	if searchText != "" {
		filter["$or"] = bson.A{
			bson.M{"title": bson.M{"$regex": searchText, "$options": "i"}},
			bson.M{"description": bson.M{"$regex": searchText, "$options": "i"}},
		}
	}

	return ms.movieRepo.GetAllMovies(ctx, filter, opts)
}

func (ms *movieService) GetShowsByLocation(ctx *gin.Context, location models.LocationMessage, movieId string) ([]models.LocationWithShow, error) {
	locations, err := ms.movieRepo.GetLocationByCoordinates(ctx, location)
	if err != nil {
		return nil, err
	}
	if len(locations) == 0 {
		return nil, nil
	}

	var locationsID []string
	for _, location := range locations {
		locationsID = append(locationsID, location.ID.Hex())
	}

	currentTime := time.Now().UTC()
	// Subtract 1 hour from current time to show ongoing shows
	currentTime = currentTime.Add(-time.Hour)

	sort := bson.M{"show_timing": 1}
	opts := options.Find().SetSort(sort)
	filter := bson.M{
		"location_id": bson.M{"$in": locationsID},
		"movie_id":    movieId,
		"end_date": bson.M{
			"$gte": currentTime,
		},
	} // Show shows for next 3 days

	shows, err := ms.movieRepo.GetShowsByFilter(ctx, filter, opts)
	if err != nil {
		return nil, err
	}

	// Process the results to include location details

	var response []models.LocationWithShow
	for _, location := range locations {
		responseShows := make([]models.Show, 0)
		if len(shows) == 0 {
			continue
		}
		for _, show := range shows {
			if show.LocationID == location.ID.Hex() {
				responseShows = append(responseShows, show)
			}
		}
		showMap := make(map[string][]models.Show)
		for _, show := range responseShows {
			// Show time is in epoch, convert to hours and minutes for display
			showTime := time.Unix(show.ShowTimingEpoch, 0).UTC()
			show.ShowTiming = showTime
			showMap[show.ShowType] = append(showMap[show.ShowType], show)
		}
		var rs []map[string][]models.Show
		rs = append(rs, showMap)
		response = append(response, models.LocationWithShow{
			Location: *location,
			Shows:    rs,
		})
	}
	if len(response) == 0 {
		return nil, errors.New("no shows found for the given location and movie")
	}
	return response, nil
}
