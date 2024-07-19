package auth

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"PaperTrail-auth.com/googleClient"
	"PaperTrail-auth.com/models"
	"PaperTrail-auth.com/utils"
	"github.com/gin-gonic/gin"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type GoogleUser struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

var googleOauthConfig = googleClient.StartCredentials()
var oauthStateString = "randomstatestring"

func HandleGoogleLogin(c *gin.Context) {
	url := googleClient.GetGoogleRedirectUrl()
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func HandleGoogleCallback(c *gin.Context) {
	if c.Query("state") != oauthStateString {
		log.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, c.Query("state"))
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	code := c.Query("code")
	googleOauthToken, err := googleClient.GetGoogleToken(googleOauthConfig, code)

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

	user := models.User{
		Email:        googleUser.Email,
		Name:         googleUser.Name,
		ID:           googleUser.ID,
		Password:     "",
		AccessToken:  googleOauthToken.AccessToken,
		RefreshToken: googleOauthToken.RefreshToken,
		TokenExpiry:  googleOauthToken.Expiry,
		Base_folder:  "",
	}

	token, err := utils.GenerateToken(user.Email, googleUser.ID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Could not authenticate user.", err)
		return
	}

	_, err = user.Save()
	if err != nil && err.Error() == "User Already created" {
		user.AccessToken = googleOauthToken.AccessToken
		user.RefreshToken = googleOauthToken.RefreshToken
		user.TokenExpiry = googleOauthToken.Expiry
		err = user.UpdateToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update token in database: " + err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"userInfo": user.GetUser(), "token": token})
		return
	}

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

func ListFiles(c *gin.Context) {

	userInfoInterface, exists := c.Get("userInfo")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "err.Error()"})
	}

	userInfo, ok := userInfoInterface.(models.UserSafe)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "err.Error()"})
	}

	user := models.User{
		Email:        userInfo.Email,
		Name:         userInfo.Name,
		ID:           userInfo.ID,
		Password:     "",
		AccessToken:  userInfo.AccessToken,
		RefreshToken: userInfo.RefreshToken,
		TokenExpiry:  userInfo.TokenExpiry,
	}

	client, err := user.GetClient(googleOauthConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	driveSrv, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve Drive client: " + err.Error()})
		return
	}

	files, err := driveSrv.Files.List().PageSize(10).Fields("nextPageToken, files(id, name)").Do()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve files: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, files.Files)
}
