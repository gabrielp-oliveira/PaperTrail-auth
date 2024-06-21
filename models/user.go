package models

import (
	"database/sql"
	"errors"
	"time"

	"PaperTrail-auth.com/db"
	"PaperTrail-auth.com/utils"
)

type User struct {
	ID       int64
	Email    string `binding:"required"`
	Password string `binding:"required"`
}

func (u User) Save() (int64, error) {
	var userID int

	query := `SELECT id FROM users WHERE email = $1`
	err := db.DB.QueryRow(query, u.Email).Scan(&userID)

	if err == sql.ErrNoRows {

		insertQuery := "INSERT INTO users(email, password, dateTime) VALUES ($1, $2, $3) RETURNING id"

		hashedPassword, err := utils.HashPassword(u.Password)

		if err != nil {
			return 0, err
		}

		// result, err := stmt.Exec(u.Email, hashedPassword)

		err = db.DB.QueryRow(insertQuery, u.Email, hashedPassword, time.Now()).Scan(&userID)
		if err != nil {
			return 0, err
		}
		return int64(userID), nil
	} else if err != nil {
		return 0, err
	} else if userID != 0 {
		return int64(userID), errors.New("User Already created")
	}
	return 0, err

}

func (u *User) ValidateCredentials() error {
	query := "SELECT id, password FROM users WHERE email = ?"
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
