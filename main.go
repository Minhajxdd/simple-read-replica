package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

var masterDb *sql.DB
var readDb *sql.DB

func main() {

	connect("postgres://root:root@localhost:5432/school?sslmode=disable", masterDb, "Master")
	connect("postgres://root:root@localhost:5433/school?sslmode=disable", readDb, "Replica")

	http.HandleFunc("/create", create)
	http.HandleFunc("/read", read)

	http.ListenAndServe(":4000", nil)
}

type Record struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "created\n")
}

func read(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "read\n")
}

func connect(uri string, db *sql.DB, name string) {
	connStr := uri
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error connecting to DB:", err)
	}
	defer db.Close()

	if err := migrate(db); err != nil {
		log.Fatal("Migration failed:", err)
	}
	fmt.Println("Connected to db : ", name)
}

func migrate(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS records (
        id SERIAL PRIMARY KEY,
        name TEXT NOT NULL
    );`
	_, err := db.Exec(query)
	return err
}
