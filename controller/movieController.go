package controller

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	apiresponse "ticketBookingSystem/internal/apiResponse"
	"ticketBookingSystem/models"
	"ticketBookingSystem/service"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type MovieController interface {
	GetAllMovies(ctx *gin.Context)
	GetMovieDetails(ctx *gin.Context)
	GetMovieShows(ctx *gin.Context)
	GetSeatingAvailability(ctx *gin.Context)
}

type movieController struct {
	movieService service.MovieService
}

func (mc *movieController) GetAllMovies(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	lastID := ctx.DefaultQuery("lastId", "")

	searchText := ctx.DefaultQuery("searchText", "")
	movies, err := mc.movieService.GetAllMovies(ctx, limit, lastID, searchText)
	if err != nil {
		apiresponse.NewErrorResponse(ctx, err, http.StatusBadRequest)
		return
	}
	nextCursor := ""
	if len(movies) != 0 {
		nextCursor = movies[len(movies)-1].ID.Hex()
	}
	resp := gin.H{"movies": movies, "nextCursor": nextCursor}
	apiresponse.NewSuccessResponse(ctx, resp, http.StatusOK)
}
func (mc *movieController) GetMovieDetails(ctx *gin.Context) {
	movieID := ctx.Param("movieId")
	if movieID == "" {
		apiresponse.NewErrorResponse(ctx, errors.New("movie id is required"), http.StatusBadRequest)
		return
	}
	movie, err := mc.movieService.GetMovieDetails(ctx, movieID)
	if err != nil {
		apiresponse.NewErrorResponse(ctx, err, http.StatusBadRequest)
		return
	}
	apiresponse.NewSuccessResponse(ctx, gin.H{"movie": movie}, http.StatusOK)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // allow all origins
}

func (mc *movieController) GetMovieShows(ctx *gin.Context) {
	movieID := ctx.Param("movieId")
	if movieID == "" {
		apiresponse.NewErrorResponse(ctx, errors.New("movie id is required"), http.StatusBadRequest)
		return
	}
	lngString := ctx.GetString("lng")
	if lngString == "" {
		apiresponse.NewErrorResponse(ctx, errors.New("location lng is required"), http.StatusBadRequest)
		return
	}
	latString := ctx.GetString("lat")
	if lngString == "" {
		apiresponse.NewErrorResponse(ctx, errors.New("location lat is required"), http.StatusBadRequest)
		return
	}
	locationInfo := &models.LocationMessage{}
	if latString != "" && lngString != "" {
		lng, err := strconv.ParseFloat(lngString, 64)
		if err != nil {
			apiresponse.NewErrorResponse(ctx, errors.New("invalid longitude"), http.StatusBadRequest)
			return
		}
		lat, err := strconv.ParseFloat(latString, 64)
		if err != nil {
			apiresponse.NewErrorResponse(ctx, errors.New("invalid latitude"), http.StatusBadRequest)
			return
		}
		locationInfo = &models.LocationMessage{
			Lat: &lat,
			Lng: &lng,
		}
	}

	limit, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	lastID := ctx.DefaultQuery("lastId", "")
	// Read messages from client

	// Fetch nearby movie shows using location message
	shows, err := mc.movieService.GetShowsByLocation(ctx, locationInfo, movieID, limit, lastID)
	if err != nil {
		log.Println("Error fetching movie shows:", err)
		apiresponse.NewErrorResponse(ctx, err, http.StatusBadRequest)
		return
	}
	apiresponse.NewSuccessResponse(ctx, shows, http.StatusOK)
}
func (mc *movieController) GetSeatingAvailability(ctx *gin.Context) {
	showId := ctx.Param("showId")
	if showId == "" {
		apiresponse.NewErrorResponse(ctx, errors.New("movie id is required"), http.StatusBadRequest)
		return
	}

	show, err := mc.movieService.GetSeatingAvailability(ctx, showId)
	if err != nil {
		log.Println("Error fetching seating availability:", err)
		apiresponse.NewErrorResponse(ctx, err, http.StatusBadRequest)
		return
	}
	apiresponse.NewSuccessResponse(ctx, show, http.StatusOK)
}

func NewMovieController(movieService service.MovieService) MovieController {
	return &movieController{
		movieService: movieService,
	}
}
