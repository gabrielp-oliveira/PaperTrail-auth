package controller

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	credentialsconfig "PaperTrail-auth.com/credentialsConfig"
	"PaperTrail-auth.com/models"
	"PaperTrail-auth.com/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

var OauthStateString = "randomstatestring"

type GoogleUser struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

var oauthStateString = "randomstatestring"

func HandleGoogleLogin(c *gin.Context) {
	url := GetGoogleRedirectUrl()
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func HandleGoogleCallback(c *gin.Context) {
	if c.Query("state") != oauthStateString {
		log.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, c.Query("state"))
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}
	var googleOauthConfig = credentialsconfig.StartGoogleCredentials()

	code := c.Query("code")
	googleOauthToken, err := GetGoogleToken(googleOauthConfig, code)

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

	var googleUser GoogleUser
	if err := json.Unmarshal(userInfo, &googleUser); err != nil {
		log.Printf("Unmarshal: %s\n", err)
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}
	token, err := utils.GenerateToken(googleUser.Email, googleUser.ID)

	user := models.User{
		Email:        googleUser.Email,
		Name:         googleUser.Name,
		ID:           googleUser.ID,
		Password:     "",
		AccessToken:  token,
		RefreshToken: googleOauthToken.RefreshToken,
		TokenExpiry:  googleOauthToken.Expiry.Add(time.Hour * 8),
		Base_folder:  "",
		Source:       "google",
	}

	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Could not authenticate user.", err)
		return
	}

	_, err = user.CreateUser(true)
	if err != nil && err.Error() == "User Already created" {
		// user.AccessToken = googleOauthToken.AccessToken
		c.Redirect(http.StatusTemporaryRedirect, "http://localhost:4200/dashboard?accessToken="+token+"&expiry="+user.TokenExpiry.Format(time.RFC3339))

		return
	}

	if err != nil {
		log.Printf("Get: %s\n", err)
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, "http://localhost:4200/dashboard?accessToken="+token+"&expiry="+user.TokenExpiry.Format(time.RFC3339))

}

func GetGoogleToken(config *oauth2.Config, code string) (*oauth2.Token, error) {
	return config.Exchange(context.Background(), code)
}

func GetGoogleRedirectUrl() string {
	return credentialsconfig.StartGoogleCredentials().AuthCodeURL(OauthStateString, oauth2.AccessTypeOffline, oauth2.ApprovalForce, oauth2.SetAuthURLParam("prompt", "consent"))
}

func GetGoogleUrl(C *gin.Context) {
	C.JSON(http.StatusOK, GetGoogleRedirectUrl())
}
