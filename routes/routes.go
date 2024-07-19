package routes

import (
	// "PaperTrail-auth.com/middlewares"

	"PaperTrail-auth.com/auth"
	"PaperTrail-auth.com/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(server *gin.Engine) {
	// server.GET("/status", status) // GET, POST, PUT, PATCH, DELETE

	logged := server.Group("/logged")
	logged.Use(middlewares.Authenticate)
	logged.POST("/listFiles", auth.ListFiles)

	server.POST("/signup", auth.Signup)
	server.POST("/login", auth.Login)

	server.GET("/", handleMain)
	server.GET("/auth/google/login", auth.HandleGoogleLogin)
	server.GET("/auth/google/callback", auth.HandleGoogleCallback)

}

// http://localhost:8080/auth/google/login
func handleMain(c *gin.Context) {
	var htmlIndex = `<html>
	<body>
	<a href="/auth/google/login">Google Log In</a>
	</body>
	</html>`

	c.Header("Content-Type", "text/html")
	c.String(200, htmlIndex)
}
