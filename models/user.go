package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	credentialsconfig "PaperTrail-auth.com/credentialsConfig"
	"PaperTrail-auth.com/db"
	"PaperTrail-auth.com/utils"
	"golang.org/x/oauth2"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email" binding:"required"`
	Name         string    `json:"name"`
	Password     string    `json:"password"`
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refresh_token"`
	TokenExpiry  time.Time `json:"token_expiry"`
	Base_folder  string    `json:"base_folder"`
	Source       string    `json:"source"`
	Verification bool      `json:"verification"`
}

type UserSafe struct {
	User
	Password string `json:"-"` // Isso omite a propriedade Password na serialização JSON
}

func (u User) GetUser() UserSafe {
	x := User{
		ID:           u.ID,
		Email:        u.Email,
		Name:         u.Name,
		AccessToken:  u.AccessToken,
		RefreshToken: u.RefreshToken,
		TokenExpiry:  u.TokenExpiry,
		Source:       u.Source,
		Verification: u.Verification,
	}
	userSafe := UserSafe{
		User: x,
	}
	return userSafe
}

func (u *User) GetUserByEmail() error {

	insertQuery := "select id, email, name, accessToken, refresh_token, token_expiry, base_folder, source, verification  from users  where email = $1"
	err := db.DB.QueryRow(insertQuery, u.Email).Scan(&u.ID, &u.Email, &u.Name, &u.AccessToken, &u.RefreshToken, &u.TokenExpiry, &u.Base_folder, &u.Source, &u.Verification)
	if err == sql.ErrNoRows {
		return errors.New("user not found")
	}
	if err != nil {
		return err
	}
	return nil
}

func (u User) ValidateUser() error {
	updateQuery := "UPDATE users SET Verification = $1 WHERE email = $2"
	_, err := db.DB.Exec(updateQuery, u.Verification, u.Email)
	if err != nil {
		return errors.New("Error updating users table. " + err.Error())
	}
	return nil
}
func (u *User) Save() (string, error) {
	var userID string

	query := `SELECT id FROM users WHERE email = $1`
	err := db.DB.QueryRow(query, u.Email).Scan(&userID)

	if err == sql.ErrNoRows {
		hashedPassword, err := utils.HashPassword(u.Password)
		if err != nil {
			return "", errors.New("hash error. " + err.Error())
		}

		var insertQuery string
		if u.ID == "" {
			insertQuery = "INSERT INTO users(email, password, created_at, name, accessToken, refresh_token, token_expiry, base_folder, source) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id"
			err = db.DB.QueryRow(insertQuery, u.Email, hashedPassword, time.Now(), u.Name, u.AccessToken, u.RefreshToken, u.TokenExpiry, u.Base_folder, u.Source).Scan(&userID)
		} else {
			insertQuery = "INSERT INTO users(email, password, created_at, id, name, accessToken, refresh_token, token_expiry, base_folder, source) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id"
			err = db.DB.QueryRow(insertQuery, u.Email, hashedPassword, time.Now(), u.ID, u.Name, u.AccessToken, u.RefreshToken, u.TokenExpiry, u.Base_folder, u.Source).Scan(&userID)
		}

		if err != nil {
			return "", errors.New("Error creating new user. " + err.Error())
		}
		return userID, nil
	} else if err != nil {
		return "", err
	} else if userID != "" {
		updateQuery := "UPDATE users SET name = $1, accessToken = $2, refresh_token = $3, token_expiry = $4 WHERE id = $5"
		_, err := db.DB.Exec(updateQuery, u.Name, u.AccessToken, u.RefreshToken, u.TokenExpiry, userID)
		if err != nil {
			return "", errors.New("Error updating user. " + err.Error())
		}
	}
	return userID, nil
}

func (u User) CreateUser(status bool) (string, error) {
	var userID string

	query := `SELECT id FROM users WHERE email = $1`
	err := db.DB.QueryRow(query, u.Email).Scan(&userID)

	if err == sql.ErrNoRows {
		hashedPassword, err := utils.HashPassword(u.Password)
		if err != nil {
			return "", errors.New("hash error. " + err.Error())
		}

		var insertQuery string
		if u.ID == "" {
			insertQuery = "INSERT INTO users(email, password, created_at, name, accessToken, refresh_token, token_expiry, base_folder, source, verification) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id"
			err = db.DB.QueryRow(insertQuery, u.Email, hashedPassword, time.Now(), u.Name, u.AccessToken, u.RefreshToken, u.TokenExpiry, u.Base_folder, u.Source, status).Scan(&userID)
		} else {
			insertQuery = "INSERT INTO users(email, password, created_at, id, name, accessToken, refresh_token, token_expiry, base_folder, source, verification) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id"
			err = db.DB.QueryRow(insertQuery, u.Email, hashedPassword, time.Now(), u.ID, u.Name, u.AccessToken, u.RefreshToken, u.TokenExpiry, u.Base_folder, u.Source, status).Scan(&userID)
		}

		if err != nil {
			return "", errors.New("Error creating new user. " + err.Error())
		}
		return userID, nil
	} else if err != nil {
		return "", err
	} else if userID != "" {
		return "", errors.New("User Already created")

	}
	return userID, nil
}

func (u User) GetClient(config *oauth2.Config) (*http.Client, error) {
	var token oauth2.Token

	// Recupere o token do banco de dados
	err := db.DB.QueryRow("SELECT accessToken, refresh_token, token_expiry FROM users WHERE email = $1", u.Email).Scan(
		&token.AccessToken, &token.RefreshToken, &token.Expiry)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from database: %v", err)
	}

	// Verifique se o token está expirado e, se necessário, use o refresh token para obter um novo token
	if time.Now().After(token.Expiry) {
		tokenSource := config.TokenSource(context.Background(), &token)
		newToken, err := tokenSource.Token()
		if err != nil {
			return nil, fmt.Errorf("unable to refresh token: %v", err)
		}
		// Atualize o token no banco de dados
		_, err = db.DB.Exec("UPDATE users SET accessToken = $1, refresh_token = $2, token_expiry = $3 WHERE email = $4",
			newToken.AccessToken, newToken.RefreshToken, newToken.Expiry, u.Email)
		if err != nil {
			return nil, fmt.Errorf("unable to update token in database: %v", err)
		}
		token = *newToken
	}

	client := config.Client(context.Background(), &token)
	return client, nil
}

func (u *User) ValidateCredentials() error {
	query := "SELECT id, password, name FROM users WHERE email = $1"
	row := db.DB.QueryRow(query, u.Email)

	var retrievedPassword string
	err := row.Scan(&u.ID, &retrievedPassword, &u.Name)

	if err != nil {
		return errors.New("invalid credentials")
	}

	passwordIsValid := utils.CheckPasswordHash(u.Password, retrievedPassword)

	if !passwordIsValid {
		return errors.New("invalid credentials")
	}

	return nil
}

func (u *User) UpdateOAuthToken() (*oauth2.Token, error) {
	var googleOauthConfig = credentialsconfig.StartGoogleCredentials()

	config := googleOauthConfig

	token := &oauth2.Token{
		AccessToken:  u.AccessToken,
		RefreshToken: u.RefreshToken,
		Expiry:       u.TokenExpiry,
	}

	tokenSource := config.TokenSource(context.Background(), token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, err
	}

	u.AccessToken = newToken.AccessToken
	u.RefreshToken = newToken.RefreshToken
	u.TokenExpiry = newToken.Expiry.Add(time.Hour * 8)

	err = u.UpdateToken()
	if err != nil {
		return nil, err
	}

	return newToken, nil
}

func (u *User) UpdateToken() error {
	updateQuery := "UPDATE users SET accessToken = $1, refresh_token = $2, token_expiry = $3 WHERE email = $4"
	_, err := db.DB.Exec(updateQuery, u.AccessToken, u.RefreshToken, u.TokenExpiry, u.Email)
	if err != nil {
		return errors.New("Error updating token. " + err.Error())
	}
	return nil
}
