package auth

import (
	"fmt"
	"net/http"
	"time"

	"PaperTrail-auth.com/emailHandler"
	"PaperTrail-auth.com/models"
	"PaperTrail-auth.com/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var oauthStateString = "randomstatestring"

func Signup(C *gin.Context) {
	var user models.User

	err := C.ShouldBindJSON(&user)
	if err != nil {
		utils.RespondWithError(C, http.StatusBadRequest, "Could not parse request data.", err)
		return
	}

	id, err := uuid.NewUUID()
	if err != nil {
		utils.RespondWithError(C, http.StatusInternalServerError, "An error occurred", err)
		return
	}
	user.ID = id.String()
	err = user.GetUserByEmail()
	if err == nil {
		utils.RespondWithError(C, http.StatusUnauthorized, "User already registered", err)
		return
	}

	user.Verification = false
	now := time.Now()

	var emailData emailHandler.EmailStruct
	hashedId, err := utils.GenerateToken(user.Email, user.ID)
	if err != nil {
		utils.RespondWithError(C, http.StatusInternalServerError, "An error occurred", err)
		return
	}

	emailData.To = append(emailData.To, user.Email)
	expiryTime := now.Add(30 * time.Minute).Format(time.RFC3339)
	link := fmt.Sprintf("http://localhost:8080/auth/validate/%s/%s", hashedId, expiryTime)
	emailData.Data = link

	_, err = emailHandler.SendEmail(C, emailData)
	if err != nil {
		utils.RespondWithError(C, http.StatusInternalServerError, "An error occurred while sending email", err)
		return
	}

	_, err = user.CreateUser(false)
	if err != nil {
		utils.RespondWithError(C, http.StatusInternalServerError, "An error occurred while saving user", err)
		return
	}

	C.JSON(http.StatusCreated, gin.H{"message": "We are good to go, please validate your email."})
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
	err = user.GetUserByEmail()
	if !user.Verification {
		utils.RespondWithError(context, http.StatusUnauthorized, "please, validate your email account.", err)
		return
	}
	if err != nil {
		utils.RespondWithError(context, http.StatusUnauthorized, "Could not get user info.", err)
		return
	}

	if !user.Verification {
		utils.RespondWithError(context, http.StatusUnauthorized, "user email not verified.", err)
		return
	}
	token, err := utils.GenerateToken(user.Email, user.ID)

	if err != nil {
		utils.RespondWithError(context, http.StatusInternalServerError, "Could not authenticate user.", err)
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Login successful!", "token": token, "userInfo": user.GetUser()})
	return
}

func ValidateEmail(C *gin.Context) {
	hashedId := C.Param("id")
	TokenExpiry := C.Param("expiryTime")

	expiryTime, err := time.Parse(time.RFC3339, TokenExpiry)
	if err != nil {
		utils.RespondWithError(C, http.StatusBadRequest, "Invalid expiry time format", err)
		return
	}

	if time.Now().After(expiryTime) {
		utils.RespondWithError(C, http.StatusBadRequest, "Token has expired", nil)
		return
	}

	email, err := utils.VerifyToken(hashedId)
	if err != nil {
		utils.RespondWithError(C, http.StatusInternalServerError, "Error validating user account", err)
		return
	}

	var user models.User
	user.Email = email
	err = user.GetUserByEmail()
	if err != nil {
		utils.RespondWithError(C, http.StatusInternalServerError, "Error retrieving user", err)
		return
	}

	if user.Verification {
		utils.RespondWithError(C, http.StatusBadRequest, "Account already verified", nil)
		return
	}

	user.Verification = true
	err = user.ValidateUser()
	if err != nil {
		utils.RespondWithError(C, http.StatusInternalServerError, "Error validating user account", err)
		return
	}

	token, err := utils.GenerateToken(user.Email, user.ID)
	if err != nil {
		utils.RespondWithError(C, http.StatusInternalServerError, "Error generating token", err)
		return
	}

	C.JSON(http.StatusOK, gin.H{"message": "Account successfully validated!", "token": token, "userInfo": user.GetUser()})
}
