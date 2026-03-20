package router

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	bookingv1 "github.com/ticketbox/pkg/proto/booking/v1"
	eventv1 "github.com/ticketbox/pkg/proto/event/v1"
	paymentv1 "github.com/ticketbox/pkg/proto/payment/v1"
	userv1 "github.com/ticketbox/pkg/proto/user/v1"

	"github.com/ticketbox/gateway/internal/handler"
	"github.com/ticketbox/gateway/internal/middleware"
)

func SetupRouter(
	userClient userv1.UserServiceClient,
	eventClient eventv1.EventServiceClient,
	bookingClient bookingv1.BookingServiceClient,
	paymentClient paymentv1.PaymentServiceClient,
	redisClient *redis.Client,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())

	authHandler := handler.NewAuthHandler(userClient, redisClient)
	eventHandler := handler.NewEventHandler(eventClient)
	bookingHandler := handler.NewBookingHandler(bookingClient)
	userHandler := handler.NewUserHandler(userClient)
	paymentHandler := handler.NewPaymentHandler(paymentClient)

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)

		api.GET("/events", eventHandler.ListEvents)
		api.GET("/events/:id", eventHandler.GetEvent)
		api.GET("/seats", eventHandler.GetSeats)

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(userClient, redisClient))
		{
			protected.POST("/auth/logout", authHandler.Logout)
			protected.GET("/users/me", userHandler.GetProfile)
			protected.PUT("/users/me", userHandler.UpdateProfile)

			protected.POST("/bookings", bookingHandler.CreateBooking)
			protected.GET("/bookings", bookingHandler.ListBookings)
			protected.GET("/bookings/:id", bookingHandler.GetBooking)
			protected.POST("/bookings/:id/cancel", bookingHandler.CancelBooking)

			protected.POST("/payments/search", paymentHandler.SearchPayments)
			protected.POST("/payments", paymentHandler.CreatePayment)
			protected.GET("/payments/:id", paymentHandler.GetPaymentById)
			protected.PATCH("/payments/:id", paymentHandler.UpdatePayment)
			protected.DELETE("/payments/:id", paymentHandler.DeletePayment)

			admin := protected.Group("")
			admin.Use(middleware.AdminOnly())
			admin.POST("/events", eventHandler.CreateEvent)
			admin.PUT("/events/:id", eventHandler.UpdateEvent)
			admin.DELETE("/events/:id", eventHandler.DeleteEvent)
		}
	}

	return r
}
