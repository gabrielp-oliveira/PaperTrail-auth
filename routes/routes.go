package routes

import (
	"PaperTrail-auth.com/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(server *gin.Engine) {
	server.GET("/status", status) // GET, POST, PUT, PATCH, DELETE

	authenticated := server.Group("/")
	authenticated.Use(middlewares.Authenticate)

	server.POST("/signup", signup)
	// server.POST("/login", login)
}
