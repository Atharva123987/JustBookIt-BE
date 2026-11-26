package service

import (
	"errors"
	"log"
	"regexp"
	"ticketBookingSystem/internal/config"
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
	SearchMovies(ctx *gin.Context, limit int, lastId string, searchText string, embedding []float32) ([]models.Movie, string, error)
	GetMovieDetails(ctx *gin.Context, movieId string) (*models.Movie, error)
	GetShowsByLocation(ctx *gin.Context, location *models.LocationMessage, movieId string, pageSize int, lastId string) ([]models.LocationWithShow, error)
	GetSeatingAvailability(ctx *gin.Context, showId string) (*models.Show, error)
}

type movieService struct {
	movieRepo     repository.MovieRepository
	knowledgeRepo repository.KnowledgeRepository
}

func NewMovieService(repo repository.MovieRepository, knowledgeRepo repository.KnowledgeRepository) MovieService {
	return &movieService{movieRepo: repo, knowledgeRepo: knowledgeRepo}
}

func (ms *movieService) GetMovieDetails(ctx *gin.Context, movieId string) (*models.Movie, error) {
	return ms.movieRepo.GetMovieById(ctx, movieId)
}

func (ms *movieService) GetAllMovies(ctx *gin.Context, limit int, lastId string, searchText string) ([]models.Movie, error) {
	if limit < 1 || limit > 100 {
		limit = 10
	}
	filter := bson.M{}
	// Sort by _id so the "_id > lastId" cursor pages through every movie exactly once.
	opts := options.Find().SetLimit(int64(limit)).SetSort(bson.M{"_id": 1})

	if lastId != "" {
		id, err := primitive.ObjectIDFromHex(lastId)
		if err != nil {
			return nil, err
		}
		filter["_id"] = bson.M{"$gt": id}
	}
	if searchText != "" {
		// Escape regex metacharacters so user input is matched literally.
		pattern := regexp.QuoteMeta(searchText)
		filter["$or"] = bson.A{
			bson.M{"title": bson.M{"$regex": pattern, "$options": "i"}},
			bson.M{"description": bson.M{"$regex": pattern, "$options": "i"}},
		}
	}

	return ms.movieRepo.GetAllMovies(ctx, filter, opts)
}

// SearchMovies returns the text search results (same as GetAllMovies) followed by
// movies found through vector search on knowledge_chunks, skipping duplicates.
// nextCursor is derived from the text search results only, since vector results
// are ordered by similarity rather than _id.
func (ms *movieService) SearchMovies(ctx *gin.Context, limit int, lastId string, searchText string, embedding []float32) ([]models.Movie, string, error) {
	if limit < 1 || limit > 100 {
		limit = 10
	}

	textMovies, err := ms.GetAllMovies(ctx, limit, lastId, searchText)
	if err != nil {
		return nil, "", err
	}

	nextCursor := ""
	if len(textMovies) != 0 {
		nextCursor = textMovies[len(textMovies)-1].ID.Hex()
	}

	if len(embedding) == 0 {
		return textMovies, nextCursor, nil
	}

	// Vector search is best-effort: if it fails, still return the text results.
	vectorMovies, err := ms.vectorSearchMovies(ctx, embedding, limit)
	if err != nil {
		log.Println("Vector search failed, returning text results only:", err)
		return textMovies, nextCursor, nil
	}

	seen := make(map[primitive.ObjectID]bool, len(textMovies))
	movies := make([]models.Movie, 0, len(textMovies)+len(vectorMovies))
	for _, movie := range textMovies {
		seen[movie.ID] = true
		movies = append(movies, movie)
	}
	for _, movie := range vectorMovies {
		if seen[movie.ID] {
			continue
		}
		seen[movie.ID] = true
		movies = append(movies, movie)
	}
	return movies, nextCursor, nil
}

// vectorSearchMovies finds the movies whose knowledge chunks are closest to the
// given embedding, ordered by similarity (best first).
func (ms *movieService) vectorSearchMovies(ctx *gin.Context, embedding []float32, limit int) ([]models.Movie, error) {
	chunks, err := ms.knowledgeRepo.VectorSearch(ctx, embedding, config.AppConfig.VectorIndexName, limit, config.AppConfig.VectorMinScore)
	if err != nil {
		return nil, err
	}

	ids := make([]primitive.ObjectID, 0, len(chunks))
	for _, chunk := range chunks {
		if chunk.Source != "" && chunk.Source != "movie" {
			continue
		}
		id, err := primitive.ObjectIDFromHex(chunk.SourceID)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return []models.Movie{}, nil
	}

	found, err := ms.movieRepo.GetAllMovies(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}

	byID := make(map[primitive.ObjectID]models.Movie, len(found))
	for _, movie := range found {
		byID[movie.ID] = movie
	}
	ordered := make([]models.Movie, 0, len(found))
	for _, id := range ids {
		if movie, ok := byID[id]; ok {
			ordered = append(ordered, movie)
			delete(byID, id)
		}
	}
	return ordered, nil
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
