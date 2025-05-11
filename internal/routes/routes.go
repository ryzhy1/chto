package routes

import (
	"boton-back/internal/handlers"
	"boton-back/internal/middlewares"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

func InitRoutes(authHandler *handlers.AuthHandler, userHandler *handlers.UserHandler, authMiddleware *middlewares.AuthMiddleware) *gin.Engine {
	r := gin.Default()

	_ = r.SetTrustedProxies(nil)

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api := r.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "ok",
			})
		})

		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/sign-in", authHandler.Login)
			//auth.POST("/refresh", accountHandler.RefreshToken)
			auth.PATCH("/email", authHandler.UpdateUserEmail)
			auth.PATCH("/password", authHandler.UpdateUserPassword)
		}

		api.Use(authMiddleware.Handle())
		{
			api.GET("/me", userHandler.GetUser)
			api.GET("/friends", userHandler.GetAllFriends)
		}
	}

	return r
}
