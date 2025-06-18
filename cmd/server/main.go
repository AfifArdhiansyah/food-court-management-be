package main

import (
	"log"

	"foodcourt-backend/internal/config"
	"foodcourt-backend/internal/database"
	"foodcourt-backend/internal/handlers"
	"foodcourt-backend/internal/middleware"
	"foodcourt-backend/pkg/auth"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// Connect to database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Run seeder
	if err := db.Seed(); err != nil {
		log.Fatalf("Failed to run seeder: %v", err)
	}

	// Initialize JWT service
	jwtService, err := auth.NewJWTService(cfg.JWT.Secret, cfg.JWT.ExpiresIn)
	if err != nil {
		log.Fatalf("Failed to initialize JWT service: %v", err)
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db.DB, jwtService)
	kiosHandler := handlers.NewKiosHandler(db.DB)
	menuHandler := handlers.NewMenuHandler(db.DB)
	orderHandler := handlers.NewOrderHandler(db.DB)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	// Initialize Gin router
	r := gin.Default()

	// Add CORS middleware
	r.Use(middleware.CORS(cfg.CORS.AllowedOrigins))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Food Court API is running",
		})
	})

	// API routes
	api := r.Group("/api/v1")

	// Auth routes (public)
	auth := api.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/register", authHandler.Register)
		auth.POST("/refresh", authHandler.RefreshToken)
	}

	// Protected routes
	protected := api.Group("/")
	protected.Use(authMiddleware.RequireAuth())
	{
		// User profile
		protected.GET("/me", authHandler.Me)

		// Kios routes
		kios := protected.Group("/kios")
		{
			kios.GET("/", kiosHandler.GetAll)
			kios.POST("/", authMiddleware.RequireRole("cashier"), kiosHandler.Create)

			// Specific kios routes with ID
			kios.GET("/:id", kiosHandler.GetByID)
			kios.PUT("/:id", authMiddleware.RequireRole("cashier"), kiosHandler.Update)
			kios.DELETE("/:id", authMiddleware.RequireRole("cashier"), kiosHandler.Delete)

			// Menu routes for specific kios
			kios.GET("/:id/menus", menuHandler.GetByKios)
			kios.POST("/:id/menus", authMiddleware.RequireRole("cashier"), menuHandler.Create)

			// Order routes for specific kios
			kios.GET("/:id/orders", orderHandler.GetByKios)
			kios.POST("/:id/orders", orderHandler.Create)
			kios.GET("/:id/queue", orderHandler.GetQueue)
		}

		// Menu routes
		menu := protected.Group("/menus")
		{
			menu.GET("/:id", menuHandler.GetByID)
			menu.PUT("/:id", menuHandler.Update)
			menu.DELETE("/:id", authMiddleware.RequireRole("cashier"), menuHandler.Delete)
		}

		// Order routes
		orders := protected.Group("/orders")
		{
			orders.GET("/", authMiddleware.RequireRole("cashier"), orderHandler.GetAll)
			orders.GET("/:id", orderHandler.GetByID)
			orders.PUT("/:id/status", orderHandler.UpdateStatus)
		}
	}

	// Start server
	log.Printf("Starting server on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
