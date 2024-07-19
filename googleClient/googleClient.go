package googleClient

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var OauthStateString = "randomstatestring"

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
			"https://www.googleapis.com/auth/drive",
			"https://www.googleapis.com/auth/documents",
		},
		Endpoint: google.Endpoint,
	}
}

func GetGoogleToken(config *oauth2.Config, code string) (*oauth2.Token, error) {
	return config.Exchange(context.Background(), code)
}

func GetGoogleRedirectUrl() string {
	return StartCredentials().AuthCodeURL(OauthStateString, oauth2.AccessTypeOffline, oauth2.ApprovalForce, oauth2.SetAuthURLParam("prompt", "consent"))
}
