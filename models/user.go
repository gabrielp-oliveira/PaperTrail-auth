package models

import (
	"database/sql"
	"errors"
	"time"

	"PaperTrail-auth.com/db"
	"PaperTrail-auth.com/utils"
)

type User struct {
	ID       int64  `json:"id"`
	Email    string `json:"email" binding:"required"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type UserWithoutPassword struct {
	ID    int64  `json:"id"`
	Email string `json:"email" binding:"required"`
	Name  string `json:"name"`
}

func (u User) GetUser() UserWithoutPassword {
	userWithoutPassword := UserWithoutPassword{
		Email: u.Email,
		Name:  u.Name,
		ID:    u.ID,
	}
	return userWithoutPassword
}

func (u User) Save() (int64, error) {
	var userID int

	query := `SELECT id FROM users WHERE email = $1`
	err := db.DB.QueryRow(query, u.Email).Scan(&userID)

	if err == sql.ErrNoRows {

		hashedPassword, err := utils.HashPassword(u.Password)

		if err != nil {
			return 0, errors.New("hash error. " + err.Error())
		}

		var insertQuery string
		if u.ID == 0 {
			insertQuery = "INSERT INTO users(email, password, created_at, name) VALUES ($1, $2, $3, $4) RETURNING id"
			err = db.DB.QueryRow(insertQuery, u.Email, hashedPassword, time.Now(), u.Name).Scan(&userID)
		} else {

			insertQuery = "INSERT INTO users(email, password, created_at, id, name) VALUES ($1, $2, $3, $4, $5) RETURNING id"
			err = db.DB.QueryRow(insertQuery, u.Email, hashedPassword, time.Now(), u.ID, u.Name).Scan(&userID)
		}

		if err != nil {
			return 0, errors.New("Error creating new user. " + err.Error())
		}
		return int64(userID), nil
	} else if err != nil {
		return 0, err
	} else if userID != 0 {
		return int64(userID), errors.New("User Already created")
	}
	return int64(userID), nil

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
