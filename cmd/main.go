package main

import (
	"log"
	"os"

	"github.com/Ulpio/guIA-backend/internal/config"
	"github.com/Ulpio/guIA-backend/internal/database"
	"github.com/Ulpio/guIA-backend/internal/handlers.go"
	"github.com/Ulpio/guIA-backend/internal/middleware"
	"github.com/Ulpio/guIA-backend/internal/repositories"
	"github.com/Ulpio/guIA-backend/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Carregar variáveis de ambiente
	if err := godotenv.Load(); err != nil {
		log.Println("Arquivo .env não encontrado, usando variáveis do sistema")
	}

	// Configurações
	cfg := config.Load()

	// Conectar ao banco de dados
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Falha ao conectar com o banco de dados:", err)
	}

	// Executar migrations
	if err := database.Migrate(db); err != nil {
		log.Fatal("Falha ao executar migrations:", err)
	}

	// Inicializar repositórios
	userRepo := repositories.NewUserRepository(db)
	postRepo := repositories.NewPostRepository(db)
	itineraryRepo := repositories.NewItineraryRepository(db)

	// Inicializar serviços
	userService := services.NewUserService(userRepo)
	postService := services.NewPostService(postRepo)
	itineraryService := services.NewItineraryService(itineraryRepo)
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	mediaService := services.NewMediaService(cfg.MediaConfig)

	// Inicializar handlers
	userHandler := handlers.NewUserHandler(userService)
	postHandler := handlers.NewPostHandler(postService)
	itineraryHandler := handlers.NewItineraryHandler(itineraryService)
	authHandler := handlers.NewAuthHandler(authService)
	mediaHandler := handlers.NewMediaHandler(mediaService)

	// Configurar Gin
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Middleware CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Rotas públicas
	api := r.Group("/api/v1")
	{
		// Autenticação
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Rotas protegidas
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			// Usuários
			users := protected.Group("/users")
			{
				users.GET("/profile", userHandler.GetProfile)
				users.PUT("/profile", userHandler.UpdateProfile)
				users.GET("/:id", userHandler.GetUserByID)
			}

			// Posts
			posts := protected.Group("/posts")
			{
				posts.GET("/", postHandler.GetFeed)
				posts.POST("/", postHandler.CreatePost)
				posts.GET("/:id", postHandler.GetPostByID)
				posts.PUT("/:id", postHandler.UpdatePost)
				posts.DELETE("/:id", postHandler.DeletePost)
				posts.POST("/:id/like", postHandler.LikePost)
				posts.DELETE("/:id/like", postHandler.UnlikePost)
			}

			// Roteiros
			itineraries := protected.Group("/itineraries")
			{
				itineraries.GET("/", itineraryHandler.GetItineraries)
				itineraries.POST("/", itineraryHandler.CreateItinerary)
				itineraries.GET("/:id", itineraryHandler.GetItineraryByID)
				itineraries.PUT("/:id", itineraryHandler.UpdateItinerary)
				itineraries.DELETE("/:id", itineraryHandler.DeleteItinerary)
				itineraries.POST("/:id/rate", itineraryHandler.RateItinerary)
			}

			// Mídia
			media := protected.Group("/media")
			{
				media.POST("/upload/image", mediaHandler.UploadImage)
				media.POST("/upload/video", mediaHandler.UploadVideo)
				media.POST("/upload/multiple", mediaHandler.UploadMultiple)
				media.DELETE("/delete", mediaHandler.DeleteMedia)
				media.GET("/info", mediaHandler.GetMediaInfo)
			}
		}
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Servir arquivos estáticos (uploads locais)
	if cfg.MediaConfig.StorageType == "local" {
		r.Static("/uploads", cfg.MediaConfig.LocalPath)
	}

	// Iniciar servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Servidor rodando na porta %s", port)
	log.Fatal(r.Run(":" + port))
}
