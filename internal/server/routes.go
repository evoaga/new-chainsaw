package server

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"new-chainsaw/internal/config"
	"new-chainsaw/internal/handlers"
	"new-chainsaw/internal/middleware"
)

func getAllowedOrigins() []string {
	if config.EnvVars["ENV"] == "production" {
		return []string{"https://your-web-url.com"}
	}
	return []string{"http://localhost:2000"}
}

func ConfigureCORS(r *gin.Engine) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     getAllowedOrigins(),
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
}

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	ConfigureCORS(r)

	r.GET("/", s.IndexHandler)

	r.GET("/hello", s.HelloWorldHandler)

	r.GET("/health", s.healthHandler)

	r.GET("/auth/:provider", handlers.AuthHandler)
	r.GET("/auth/:provider/callback", handlers.AuthCallback)

	r.GET("/user/:username", handlers.GetUserProfileByUsernameHandler)

	protected := r.Group("/")
	protected.Use(middleware.JWTMiddleware(s.dbPool))
	{
		protected.GET("/search", handlers.SearchUsers)

		protected.GET("/protected-endpoint", protectedEndpointHandler)

		protected.GET("/session", handlers.SessionHandler)
		protected.POST("/sign-out", handlers.SignOutHandler)
		protected.DELETE("/delete-account", handlers.DeleteAccountHandler)
		protected.GET("/check-username", handlers.CheckUsernameAvailabilityHandler)
		protected.PATCH("/update-user", handlers.UpdateUserHandler)

		/* */
		protected.GET("/user/profile", handlers.GetUserProfileByIDHandler)
		/* */

		protected.POST("/log-exercises", handlers.LogExerciseHandler)
		protected.GET("/exercises/latest", handlers.GetLatestExercises)

		protected.POST("/validate-save-trophies", handlers.ValidateAndSaveTrophiesHandler)
		protected.GET("/trophies", handlers.GetTrophiesHandler)
		protected.DELETE("/trophies/:display_order", handlers.DeleteTrophy)

	}

	return r
}

func (s *Server) IndexHandler(c *gin.Context) {
	c.File("./web/index.html")
}

func (s *Server) HelloWorldHandler(c *gin.Context) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}

func protectedEndpointHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "You have accessed a protected endpoint!"})
}
