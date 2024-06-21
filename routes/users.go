package routes

import (
	"net/http"

	"PaperTrail-auth.com/models"
	"PaperTrail-auth.com/utils"

	"github.com/gin-gonic/gin"
)

func signup(context *gin.Context) {
	var user models.User

	err := context.ShouldBindJSON(&user)

	if err != nil {
		utils.RespondWithError(context, http.StatusBadRequest, "Could not parse request data.", err)
		return
	}

	userId, err := user.Save()
	if err != nil {
		utils.RespondWithError(context, http.StatusInternalServerError, "An error occurred", err)
		return
	}
	user.ID = userId
	err = createInitialPapper(&user)
	if err != nil {
		utils.RespondWithError(context, http.StatusBadRequest, "Could not create initial papper.", err)
		return
	}

	context.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})

}
