package routes

import (
	// "PaperTrail-auth.com/middlewares"

	"PaperTrail-auth.com/auth"
	"PaperTrail-auth.com/googleClient"
	"PaperTrail-auth.com/microsoftClient"
	"PaperTrail-auth.com/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(server *gin.Engine) {
	// server.GET("/status", status) // GET, POST, PUT, PATCH, DELETE

	logged := server.Group("/logged")
	logged.Use(middlewares.Authenticate)

	server.POST("/signup", auth.Signup)
	server.POST("/login", auth.Login)

	server.GET("/", handleMain)
	server.GET("/auth/google/login", googleClient.HandleGoogleLogin)
	server.GET("/auth/google/callback", googleClient.HandleGoogleCallback)

	server.GET("/auth/microsoft/login", microsoftClient.HandleMicrosoftLogin)
	server.GET("/auth/microsoft/callback", microsoftClient.HandleMicrosoftCallback)
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
