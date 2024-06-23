package auth

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"PaperTrail-auth.com/models"
	"PaperTrail-auth.com/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleUser struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

var (
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:8080/auth/google/callback",
		ClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
	oauthStateString = "randomstatestring"
)

func HandleGoogleLogin(c *gin.Context) {
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func HandleGoogleCallback(c *gin.Context) {
	if c.Query("state") != oauthStateString {
		log.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, c.Query("state"))
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	code := c.Query("code")

	googleOauthToken, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + googleOauthToken.AccessToken)
	if err != nil {
		log.Printf("Get: %s\n", err)
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}
	defer response.Body.Close()

	userInfo, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("ReadAll: %s\n", err)
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	log.Printf("UserInfo: %s\n", userInfo)
	var googleUser GoogleUser
	if err := json.Unmarshal(userInfo, &googleUser); err != nil {
		log.Printf("Unmarshal: %s\n", err)
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}
	i, _ := strconv.ParseInt(googleUser.ID, 10, 64)

	user := models.User{
		Email: googleUser.Email,
		ID:    i,
		// Como a senha não é fornecida pelo Google, você pode deixar em branco ou gerar uma senha temporária/aleatória
		// Password: "",
	}

	token, err := utils.GenerateToken(user.Email, user.ID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Could not authenticate user.", err)
		return
	}
	_, err = user.Save()

	if err != nil && err.Error() == "User Already created" {
		c.JSON(http.StatusCreated, gin.H{"userInfo": user.GetUser(), "token": token})
		return
	}

	if err != nil {
		log.Printf("Get: %s\n", err)
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}
	err = user.CreateInitialPapper()
	if err != nil {
		log.Printf("Get: %s\n", err)
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully", "userInfo": user.GetUser(), "token": token})
}

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
	err = user.CreateInitialPapper()
	if err != nil {
		utils.RespondWithError(context, http.StatusBadRequest, "Could not create initial papper.", err)
		return
	}

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

	context.JSON(http.StatusOK, gin.H{"message": "Login successful!", "token": token})
}
