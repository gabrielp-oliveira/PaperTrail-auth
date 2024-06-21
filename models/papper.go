package models

import (
	"time"

	"PaperTrail-auth.com/db"
)

type Papper struct {
	ID          int64
	Name        string    `binding:"required"`
	Description string    `binding:"required"`
	Path        string    `binding:"required"`
	DateTime    time.Time `binding:"required"`
	UserID      int64     `binding:"required"`
}

func (e *Papper) Save() error {
	query := `
	INSERT INTO pappers(name, description, path, dateTime, user_id) 
	VALUES ($1, $2, $3, $4, $5) `
	stmt, err := db.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(e.Name, e.Description, e.Path, e.DateTime, e.UserID)
	if err != nil {
		return err
	}

	return err
}

func GetPapperByID(id int64) (*Papper, error) {
	query := "SELECT * FROM pappers WHERE id = ?"
	row := db.DB.QueryRow(query, id)

	var papper Papper
	err := row.Scan(&papper.ID, &papper.Name, &papper.Description, &papper.Path, &papper.DateTime, &papper.UserID)
	if err != nil {
		return nil, err
	}

	return &papper, nil
}
