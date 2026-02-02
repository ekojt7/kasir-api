package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	conn := os.Getenv("DB_CONN")
	log.Println("DB_CONN =", conn)

	if conn == "" {
		log.Fatal("DB_CONN is EMPTY")
	}

	db, err := sql.Open("postgres", conn)
	if err != nil {
		log.Fatal("OPEN ERROR:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("PING ERROR:", err)
	}

	log.Println("✅ DATABASE CONNECTED SUCCESSFULLY")
}
