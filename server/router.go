package router

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

	config.InitConfig()

	// Database init
	mongoDbAdapter := NewMongoDbAdapter(context.Background(), config.AppConfig.MongoURI)
	if mongoDbAdapter == nil {
		panic("Database not initialized")
	}
	defer func() {
		if err := mongoDbAdapter.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	// Panic recovery middleware
	r.Use(gin.Recovery())

	// Middleware
	userAuth := middleware.AuthMiddleware
	anonMiddleware := middleware.AnonymousMiddleware
	r.Use(anonMiddleware())

	movieRepository := repository.NewMovieRepository(mongoDbAdapter)
	bookingRepository := repository.NewBookingRepository(mongoDbAdapter, movieRepository)
	userStateRepository := repository.NewUserStateRepository(mongoDbAdapter)

	movieService := service.NewMovieService(movieRepository)
	paymentService := service.NewPaymentService()
	bookingService := service.NewBookingService(bookingRepository, paymentService)
	userStateService := service.NewUserStateService(userStateRepository)

	// Init providers
	movieController := controller.NewMovieController(movieService)
	bookingController := controller.NewBookingController(bookingService)
	ticketController := controller.NewTicketController()
	webhookController := controller.NewWebhookController(bookingService)
	userStateController := controller.NewUserStateController(userStateService)

	// Routes
	movieRoute := r.Group("movie")
	{
		movieRoute.GET("", movieController.GetAllMovies)                        // Get movies by search term and pagination
		movieRoute.GET(":movieId", movieController.GetMovieDetails)             // Get movie details like title, desc, cast etc.
		movieRoute.GET(":movieId/shows", movieController.GetMovieShows)         // Gets the nearby shows according to location
		movieRoute.GET("seats/:showId", movieController.GetSeatingAvailability) // Gets the seating availability for the show. For ex- 12pm show at Inorbit

		bookRoute := movieRoute.Group("book")
		{
			bookRoute.GET("", userAuth(), bookingController.GetBookings)                      // Get bookings for the user
			bookRoute.GET("/:bookingId", userAuth(), bookingController.GetBookingByBookingId) // Get bookings for the user
			bookRoute.POST("", userAuth(), bookingController.ReserveMovieSeat)                // This will hold the seat and initialize payment and return the booking id

		}
	}

	ticketRoute := r.Group("/ticket")
	{
		ticketRoute.GET("/:bookingId", userAuth(), ticketController.GetTicket) // Get ticket by booking id, it should be rotating QR / rotating id (websocket)
		ticketRoute.POST("", userAuth(), ticketController.VerifyTicket)        // Verify the ticket
	}

	userStateRoute := r.Group("/state")
	{
		userStateRoute.GET("/", userAuth(), userStateController.GetState)
		userStateRoute.POST("/", userAuth(), userStateController.UpdateState)
		userStateRoute.DELETE("/intent", userAuth(), userStateController.ClearIntent)
	}

	webhookRoute := r.Group("/webhook")
	{
		webhookRoute.POST("/stripe", webhookController.StripeWebhook)
	}

	if err := r.Run(":" + config.AppConfig.Port); err != nil {
		log.Fatal(err)
	}
	log.Println("server running at :", config.AppConfig.Port)
}
