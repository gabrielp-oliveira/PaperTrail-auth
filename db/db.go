package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "myuser"     // Substitua pelo seu usuário
	password = "mypassword" // Substitua pela sua senha
	dbname   = "mydatabase" // Substitua pelo seu banco de dados
)

var DB *sql.DB

func InitDB() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	createTables(DB)
}

func createTables(db *sql.DB) {
	createUsersTable := `
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    accessToken TEXT,
    refresh_token TEXT,
    base_folder TEXT,
	source TEXT,
    token_expiry TIMESTAMP
);
`

	_, err := db.Exec(createUsersTable)
	if err != nil {
		log.Fatalf("Could not create users table: %v", err)
	}

}
