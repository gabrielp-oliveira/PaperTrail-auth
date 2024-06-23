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
	Password string `json:"password" binding:"required"`
}

type UserWithoutPassword struct {
	Email string `json:"email" binding:"required"`
	User
}

func (u User) GetUser() UserWithoutPassword {
	userWithoutPassword := UserWithoutPassword{
		Email: u.Email,
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
			return 0, err
		}

		var insertQuery string
		if u.ID == 0 {
			insertQuery = "INSERT INTO users(email, password, created_at) VALUES ($1, $2, $3) RETURNING id"
			err = db.DB.QueryRow(insertQuery, u.Email, hashedPassword, time.Now()).Scan(&userID)
		} else {

			insertQuery = "INSERT INTO users(email, password, created_at, id) VALUES ($1, $2, $3, $4) RETURNING id"
			err = db.DB.QueryRow(insertQuery, u.Email, hashedPassword, time.Now(), u.ID).Scan(&userID)
		}

		if err != nil {
			return 0, err
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
	query := "SELECT id, password FROM users WHERE email = $1"
	row := db.DB.QueryRow(query, u.Email)

	var retrievedPassword string
	err := row.Scan(&u.ID, &retrievedPassword)

	if err != nil {
		return errors.New("credentials invalid")
	}

	passwordIsValid := utils.CheckPasswordHash(u.Password, retrievedPassword)

	if !passwordIsValid {
		return errors.New("credentials invalid")
	}

	return nil
}

func (u *User) CreateInitialPapper() error {
	var papper Papper

	var path string = u.Email + "/new papper"

	papper.Name = "new Papper"
	papper.Description = "Default new papper description"
	papper.Path = path
	papper.DateTime = time.Now()
	papper.UserID = u.ID

	err := papper.Save()
	if err != nil {
		return err
	}
	return nil
}
