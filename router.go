package main

import (
	"context"
	"log"
	"ticketBookingSystem/controller"
	"ticketBookingSystem/internal/config"
	"ticketBookingSystem/middleware"
	"ticketBookingSystem/repository"
	"ticketBookingSystem/service"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewMongoDbAdapter(ctx context.Context, uri string) *mongo.Client {
	client, err := mongo.Connect(ctx, options.Client().
		ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	return client
}

func NewRouter() {
	r := gin.Default()

	cfg := config.LoadConfig()

	// Database init
	mongoDbAdapter := NewMongoDbAdapter(context.Background(), cfg.MongoURI)
	if mongoDbAdapter == nil {
		panic("Database not initialized")
	}
	defer func() {
		if err := mongoDbAdapter.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	// Middleware
	auth := middleware.AuthMiddleware

	movieRepository := repository.NewMovieRepository(mongoDbAdapter)

	movieService := service.NewMovieService(movieRepository)

	// Init providers
	movieController := controller.NewMovieController(movieService)
	bookingController := controller.NewBookingController()
	ticketController := controller.NewTicketController()

	// Routes
	movieRoute := r.Group("movie")
	{
		movieRoute.GET("/", movieController.GetAllMovies)                        // Get movies by search term and pagination
		movieRoute.GET("/:movieId", movieController.GetMovieDetails)             // Get movie details like title, desc, cast etc.
		movieRoute.GET("/:movieId/shows", movieController.GetMovieShows)         // Gets the nearby shows according to location
		movieRoute.GET("/seats/:showId", movieController.GetSeatingAvailability) // Gets the seating availability for the show. For ex- 12pm show at Inorbit
	}

	bookRoute := movieRoute.Group("/book")
	{
		bookRoute.GET("/", auth(), bookingController.GetBookings)                        // Get bookings for the user
		bookRoute.POST("/:showId", auth(), bookingController.ReserveMovieSeat)           // This will hold the seat and initialize payment and return the booking id
		bookRoute.POST("/status/:bookingId", auth(), bookingController.GetBookingStatus) // This will update the payment status (websocket)
	}

	ticketRoute := r.Group("/ticket")
	{
		ticketRoute.GET("/:bookingId", ticketController.GetTicket) // Get ticket by booking id, it should be rotating QR / rotating id (websocket)
	}
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
	log.Println("server running at :", cfg.Port)
}
