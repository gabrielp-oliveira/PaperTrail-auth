package credentialsconfig

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/microsoft"
)

func StartMicrosoftCredentials() *oauth2.Config {
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

func StartGoogleCredentials() *oauth2.Config {
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
