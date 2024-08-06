package auth

import (
	"net/http"

	"PaperTrail-auth.com/models"
	"PaperTrail-auth.com/utils"
	"github.com/gin-gonic/gin"
)

var oauthStateString = "randomstatestring"

func Signup(context *gin.Context) {
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
	// mandar email para validar o email.

	context.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

func Login(context *gin.Context) {
	var user models.User

	err := context.ShouldBindJSON(&user)

	if err != nil {
		utils.RespondWithError(context, http.StatusBadRequest, "Could not parse request data.", err)
		return
	}

	err = user.ValidateCredentials()

	if err != nil {
		utils.RespondWithError(context, http.StatusUnauthorized, "Could not authenticate user.", err)
		return
	}

	token, err := utils.GenerateToken(user.Email, user.ID)

	if err != nil {
		utils.RespondWithError(context, http.StatusInternalServerError, "Could not authenticate user.", err)
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Login successful!", "token": token, "userInfo": user.GetUser()})
}
