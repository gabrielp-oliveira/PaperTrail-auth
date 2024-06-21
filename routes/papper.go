package routes

import (
	"net/http"
	"time"

	"PaperTrail-auth.com/models"
	"github.com/gin-gonic/gin"
)

func status(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{"message": "server up and running."})
}

func createInitialPapper(u *models.User) error {
	var papper models.Papper

	var path string = u.Email + "/new papper"

	papper.Name = "new Papper"
	papper.Description = "Default new papper description"
	papper.Path = path
	papper.DateTime = time.Now()
	papper.UserID = u.ID

	err := papper.Save()
	if err != nil {
		return err
	}
	return nil
}
