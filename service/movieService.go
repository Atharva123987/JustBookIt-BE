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
	GetShowsByLocation(ctx *gin.Context, location *models.LocationMessage, movieId string, pageSize int, lastId string) ([]models.LocationWithShow, error)
	GetSeatingAvailability(ctx *gin.Context, showId string) (*models.Show, error)
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
	opts := options.Find().SetLimit(int64(limit))

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

func (ms *movieService) GetShowsByLocation(ctx *gin.Context, location *models.LocationMessage, movieId string, pageSize int, lastId string) ([]models.LocationWithShow, error) {
	var locationsID []string
	var locations []*models.Location
	var err error
	if location.Lat != nil && location.Lng != nil {
		locations, err = ms.movieRepo.GetLocationByCoordinates(ctx, location, int64(pageSize), lastId)
		if err != nil {
			return nil, err
		}
		if len(locations) == 0 {
			return nil, nil
		}
		for _, location := range locations {
			locationsID = append(locationsID, location.ID.Hex())
		}
	}

	currentTimeEpoch := time.Now().Add(-time.Hour).UTC().Unix()

	sort := bson.M{"show_timing": 1}
	proj := bson.M{"_id": 1, "movie_id": 1, "movie_title": 1, "screen_number": 1, "show_timing_epoch": 1, "show_type": 1, "location_id": 1, "show_timing": 1, "price_label": 1}
	opts := options.Find().SetSort(sort).SetProjection(proj)
	filter := bson.M{
		"movie_id": movieId,
		"show_timing_epoch": bson.M{
			"$gte": currentTimeEpoch,
		},
	} // Show shows for next 3 days
	if len(locationsID) != 0 {
		filter["location_id"] = bson.M{"$in": locationsID}
	}
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
		if len(responseShows) == 0 {
			continue
		}
		showMap := make(map[string][]models.Show)
		for _, show := range responseShows {
			// Show time is in epoch, convert to hours and minutes for display
			showTime := time.Unix(show.ShowTimingEpoch, 0).UTC()
			show.ShowTiming = &showTime
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

func (ms *movieService) GetSeatingAvailability(ctx *gin.Context, showId string) (*models.Show, error) {
	return ms.movieRepo.GetShowById(ctx, showId)
}
