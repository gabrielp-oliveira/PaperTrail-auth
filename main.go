package main

import (
	"PaperTrail-auth.com/db"
	"PaperTrail-auth.com/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	db.InitDB()
	server := gin.Default()

	routes.RegisterRoutes(server)

	server.Run(":8080") // localhost:8080
}
