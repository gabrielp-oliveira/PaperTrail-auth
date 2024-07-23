package middlewares

import (
	"net/http"
	"time"

	"PaperTrail-auth.com/db"
	"PaperTrail-auth.com/models"
	"PaperTrail-auth.com/utils"

	"github.com/gin-gonic/gin"
)

func Authenticate(context *gin.Context) {
	token := context.Request.Header.Get("Authorization")

	if token == "" {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Not authorized."})
		return
	}

	userEmail, err := utils.VerifyToken(token)

	if err != nil {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Not authorized."})
		return
	}

	query := "SELECT name, email, id FROM users WHERE email = $1"
	row := db.DB.QueryRow(query, userEmail)

	var userInfo models.UserSafe
	err = row.Scan(&userInfo.Name, &userInfo.Email, &userInfo.ID)

	if err != nil {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Not authorized."})
	}

	
	context.Writer.Header().Set("Access-Token", userInfo.AccessToken)
	context.Writer.Header().Set("Refresh-Token", userInfo.RefreshToken)
	context.Writer.Header().Set("Token-Expiry", userInfo.TokenExpiry.Format(time.RFC3339))


	context.Set("userInfo", userInfo)

	context.Next()
}
