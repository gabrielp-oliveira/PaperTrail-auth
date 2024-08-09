package routes

import (
	// "PaperTrail-auth.com/middlewares"

	"PaperTrail-auth.com/controller"
	"PaperTrail-auth.com/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(server *gin.Engine) {
	// server.GET("/status", status) // GET, POST, PUT, PATCH, DELETE

	logged := server.Group("/logged")
	logged.Use(middlewares.Authenticate)

	server.POST("/signup", controller.Signup)
	server.POST("/login", controller.Login)
	server.GET("/auth/validate/:id/:expiryTime", controller.ValidateEmail)

	server.GET("/", handleMain)
	server.GET("/auth/google/login", controller.HandleGoogleLogin)
	server.GET("/auth/google/getUrl", controller.GetGoogleUrl)
	server.GET("/auth/google/callback", controller.HandleGoogleCallback)

	server.GET("/auth/microsoft/login", controller.HandleMicrosoftLogin)
	server.GET("/auth/microsoft/getUrl", controller.GetMicrosoftUrl)
	server.GET("/auth/microsoft/callback", controller.HandleMicrosoftCallback)
}

// http://localhost:8080/auth/google/login
func handleMain(c *gin.Context) {
	var htmlIndex = `<html>
	<body>
	<a href="/auth/google/login">Google Log In</a>
	<br/>
	<a href="/auth/microsoft/login">microsoft Log In</a>
	</body>
	</html>`

	c.Header("Content-Type", "text/html")
	c.String(200, htmlIndex)
}
