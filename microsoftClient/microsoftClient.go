package microsoftClient

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"PaperTrail-auth.com/models"
	"PaperTrail-auth.com/utils"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
)

type MicrosoftUser struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// StartCredentials initializes the OAuth2 configuration using environment variables
func StartCredentials() *oauth2.Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf(".env load error: %v", err)
	}
	ClientID := os.Getenv("MICROSOFT_CLIENT_ID")
	if ClientID == "" {
		log.Fatalf("credentials error: MICROSOFT_CLIENT_ID is missing in local env variables.")
	}

	ClientSecret := os.Getenv("MICROSOFT_SECRET_VALUE")
	if ClientSecret == "" {
		log.Fatalf("credentials error: MICROSOFT_SECRET_VALUE is missing in local env variables.")
	}
	TenantId := os.Getenv("MICROSOFT_TENANT_ID")
	if TenantId == "" {
		log.Fatalf("credentials error: MICROSOFT_TENANT_ID is missing in local env variables.")
	}
	// permissionId := os.Getenv("MICROSOFT_PERMISSION_ID")
	// if TenantId == "" {
	// 	log.Fatalf("credentials error: MICROSOFT_PERMISSION_ID is missing in local env variables.")
	// }

	return &oauth2.Config{
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
		RedirectURL:  "http://localhost:8080/auth/microsoft/callback",
		Scopes:       []string{"https://graph.microsoft.com/User.Read", "https://graph.microsoft.com/User.ReadBasic.All"},
		Endpoint:     microsoft.AzureADEndpoint("common"),
	}

}

func getstateString() string {
	stateString := os.Getenv("RANDOM_STATE_STRING")
	if stateString == "" {
		log.Fatalf("credentials error: RANDOM_STATE_STRING is missing in local env variables.")
	}
	return stateString
}

// HandleMicrosoftLogin redirects the user to the Microsoft login page
func HandleMicrosoftLogin(c *gin.Context) {
	credentials := StartCredentials()
	url := credentials.AuthCodeURL(getstateString())
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func GetMicrosoftUrl(C *gin.Context) {
	credentials := StartCredentials()
	C.JSON(http.StatusOK, credentials.AuthCodeURL(getstateString()))
}

// HandleMicrosoftCallback handles the callback from Microsoft and retrieves the user info
func HandleMicrosoftCallback(C *gin.Context) {
	if C.Query("state") != getstateString() {
		log.Printf("invalid oauth state, expected '%s', got '%s'\n", getstateString(), C.Query("state"))
		C.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	code := C.Query("code")
	if code == "" {
		log.Println("Code not found")
		C.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	credentials := StartCredentials()
	microsoftOauthToken, err := credentials.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("oauthConfig.Exchange() failed with '%s'\n", err)
		C.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	client := credentials.Client(context.Background(), microsoftOauthToken)
	resp, err := client.Get("https://graph.microsoft.com/v1.0/me")
	if err != nil {
		log.Printf("client.Get() failed with '%s'\n", err)
		C.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		log.Printf("json.NewDecoder().Decode() failed with '%s'\n", err)
		C.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	log.Printf("Unmarshal: %s\n", userInfo["mail"])
	log.Printf("Unmarshal: %s\n", userInfo["userPrincipalName"])

	token, err := utils.GenerateToken(userInfo["mail"].(string), userInfo["id"].(string))
	if err != nil {
		log.Printf("Unmarshal: %s\n", err)
		C.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	user := models.User{
		Email:        userInfo["mail"].(string),
		Name:         userInfo["displayName"].(string),
		ID:           userInfo["id"].(string),
		Password:     "",
		AccessToken:  token,
		RefreshToken: microsoftOauthToken.RefreshToken,
		TokenExpiry:  microsoftOauthToken.Expiry.Add(4 * time.Hour),
		Base_folder:  "",
		Source:       "microsoft",
	}
	_, err = user.CreateUser(true)
	if err != nil && err.Error() == "User Already created" {
		// user.AccessToken = googleOauthToken.AccessToken
		C.Redirect(http.StatusTemporaryRedirect, "http://localhost:4200/dashboard?accessToken="+token+"&expiry="+user.TokenExpiry.Format(time.RFC3339))

		return
	}
	if err != nil {
		log.Printf("Get: %s\n", err)
		C.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	redirectURL := fmt.Sprintf("http://localhost:4200/dashboard?accessToken=%s&expiry=%s", token, user.TokenExpiry)
	C.Redirect(http.StatusTemporaryRedirect, redirectURL)
}
