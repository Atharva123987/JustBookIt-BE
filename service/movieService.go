package service

import (
	"errors"
	"log"
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
	// GetAllMovies returns the text-search results followed by the vector-search
	// results (when embeddings are given), plus the cursor for the next page.
	GetAllMovies(ctx *gin.Context, limit int, lastId string, searchText string, embeddings []float32) ([]models.Movie, string, error)
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

func (ms *movieService) GetAllMovies(ctx *gin.Context, limit int, lastId string, searchText string, embeddings []float32) ([]models.Movie, string, error) {
	if limit < 1 || limit > 100 {
		limit = 10
	}
	filter := bson.M{}
	opts := options.Find().SetLimit(int64(limit))

	if lastId != "" {
		id, err := primitive.ObjectIDFromHex(lastId)
		if err != nil {
			return nil, "", err
		}
		filter["_id"] = bson.M{"$gt": id}
	}
	if searchText != "" {
		filter["$or"] = bson.A{
			bson.M{"title": bson.M{"$regex": searchText, "$options": "i"}},
			bson.M{"description": bson.M{"$regex": searchText, "$options": "i"}},
		}
	}

	movies, err := ms.movieRepo.GetAllMovies(ctx, filter, opts)
	if err != nil {
		return nil, "", err
	}

	// The cursor only follows the text-search results, so paging works as before.
	nextCursor := ""
	if len(movies) != 0 {
		nextCursor = movies[len(movies)-1].ID.Hex()
	}

	// Vector results are only added to the first page; later pages would
	// otherwise repeat them.
	if len(embeddings) == 0 || lastId != "" {
		return movies, nextCursor, nil
	}

	similar, err := ms.searchMoviesByEmbedding(ctx, embeddings, movies)
	if err != nil {
		// Vector search is best effort: keep the text results if it fails
		// (for example when the Atlas vector index does not exist yet).
		log.Println("Vector search failed, returning text search results only:", err)
		return movies, nextCursor, nil
	}

	return append(movies, similar...), nextCursor, nil
}

// searchMoviesByEmbedding finds movies whose knowledge chunks are closest to the
// given embedding, skipping movies already in exclude. Results are ordered by
// similarity, most similar first.
func (ms *movieService) searchMoviesByEmbedding(ctx *gin.Context, embeddings []float32, exclude []models.Movie) ([]models.Movie, error) {
	excludeIDs := make([]string, 0, len(exclude))
	for _, movie := range exclude {
		excludeIDs = append(excludeIDs, movie.ID.Hex())
	}

	chunks, err := ms.knowledgeRepo.VectorSearchChunks(ctx, repository.VectorSearchParams{
		IndexName:        config.AppConfig.VectorSearchIndex,
		QueryVector:      embeddings,
		Limit:            config.AppConfig.VectorSearchLimit,
		NumCandidates:    config.AppConfig.VectorSearchNumCandidates,
		MinScore:         config.AppConfig.VectorSearchMinScore,
		ExcludeSourceIDs: excludeIDs,
	})
	if err != nil {
		return nil, err
	}

	// Collect movie ids in similarity order; one movie may have several chunks.
	seen := make(map[string]bool, len(chunks))
	orderedIDs := make([]primitive.ObjectID, 0, len(chunks))
	for _, chunk := range chunks {
		if seen[chunk.SourceID] {
			continue
		}
		id, err := primitive.ObjectIDFromHex(chunk.SourceID)
		if err != nil {
			log.Println("Skipping knowledge chunk with invalid source_id:", chunk.SourceID)
			continue
		}
		seen[chunk.SourceID] = true
		orderedIDs = append(orderedIDs, id)
	}
	if len(orderedIDs) == 0 {
		return []models.Movie{}, nil
	}

	found, err := ms.movieRepo.GetAllMovies(ctx, bson.M{"_id": bson.M{"$in": orderedIDs}})
	if err != nil {
		return nil, err
	}

	// $in does not keep order, so restore the similarity order.
	byID := make(map[primitive.ObjectID]models.Movie, len(found))
	for _, movie := range found {
		byID[movie.ID] = movie
	}
	result := make([]models.Movie, 0, len(found))
	for _, id := range orderedIDs {
		if movie, ok := byID[id]; ok {
			result = append(result, movie)
		}
	}
	return result, nil
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
