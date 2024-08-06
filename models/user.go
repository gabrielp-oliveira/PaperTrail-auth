package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

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
	}
	userSafe := UserSafe{
		User: x,
	}
	return userSafe
}

func (u User) Save() (string, error) {
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
		updateQuery := "UPDATE users SET name = $1, password = $2, accessToken = $3, refresh_token = $4, token_expiry = $5 WHERE id = $6"
		_, err := db.DB.Exec(updateQuery, u.Name, u.Password, u.AccessToken, u.RefreshToken, u.TokenExpiry, userID)
		if err != nil {
			return "", errors.New("Error updating user. " + err.Error())
		}
	}
	return userID, nil
}

func (u User) CreateUser() (string, error) {
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

func (u User) SetToken() error {
	insertQuery := "INSERT INTO users(email, password, created_at, id, name, accessToken, refresh_token, token_expiry) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id"

	_, err := db.DB.Exec(insertQuery,
		u.Email, "", time.Now(), u.ID, u.Name, u.AccessToken, u.RefreshToken, u.TokenExpiry)
	if err != nil {
		return errors.New("Unable to save token in database: " + err.Error())

	}
	return nil
}
func (u *User) ValidateCredentials() error {
	query := "SELECT id, password, name FROM users WHERE email = $1"
	row := db.DB.QueryRow(query, u.Email)

	var retrievedPassword string
	err := row.Scan(&u.ID, &retrievedPassword, &u.Name)

	if err != nil {
		return errors.New("credenciais inválidas")
	}

	passwordIsValid := utils.CheckPasswordHash(u.Password, retrievedPassword)

	if !passwordIsValid {
		return errors.New("credentials invalid")
	}

	return nil
}
