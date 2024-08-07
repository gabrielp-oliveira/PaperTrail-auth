package googleClient

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"PaperTrail-auth.com/models"
	"PaperTrail-auth.com/utils"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var OauthStateString = "randomstatestring"

type GoogleUser struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func StartCredentials() *oauth2.Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf(".env load error: %v", err)
	}
	ClientID := os.Getenv("GOOGLE_OAUTH_CLIENT_ID")
	if ClientID == "" {
		log.Fatalf("credentials error: %v", err)
	}

	ClientSecret := os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET")
	if ClientSecret == "" {
		log.Fatalf("credentials error: %v", err)
	}

	return &oauth2.Config{
		RedirectURL:  "http://localhost:8080/auth/google/callback",
		ClientSecret: ClientSecret,
		ClientID:     ClientID,
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}
}

var googleOauthConfig = StartCredentials()
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
	return StartCredentials().AuthCodeURL(OauthStateString, oauth2.AccessTypeOffline, oauth2.ApprovalForce, oauth2.SetAuthURLParam("prompt", "consent"))
}

func GetGoogleUrl(C *gin.Context) {
	C.JSON(http.StatusOK, GetGoogleRedirectUrl())
}
