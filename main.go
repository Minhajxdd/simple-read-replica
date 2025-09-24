package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

type Record struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var masterDb *sql.DB
var readDb *sql.DB

func main() {
	var err error

	masterDb, err = connect("postgres://root:root@localhost:5432/school?sslmode=disable", "Master")
	if err != nil {
		log.Fatal(err)
	}

	if err := migrate(masterDb); err != nil {
		log.Fatal("migration failed for master cluster: ", err)
	}

	readDb, err = connect("postgres://root:root@localhost:5433/school?sslmode=disable", "Replica")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/read/", readHandler)

	fmt.Println("Server running on :4000")
	log.Fatal(http.ListenAndServe(":4000", nil))
}

func connect(uri string, name string) (*sql.DB, error) {
	db, err := sql.Open("postgres", uri)
	if err != nil {
		return nil, fmt.Errorf("error connecting to %s: %v", name, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping failed for %s: %v", name, err)
	}

	fmt.Println("Connected to DB:", name)
	return db, nil
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

func createHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var id int
	err := masterDb.QueryRow("INSERT INTO records (name) VALUES ($1) RETURNING id", input.Name).Scan(&id)
	if err != nil {
		http.Error(w, "DB insert failed", http.StatusInternalServerError)
		return
	}

	record := Record{ID: id, Name: input.Name}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(record)
}

func readHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/read/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var record Record
	err = readDb.QueryRow("SELECT id, name FROM records WHERE id=$1", id).Scan(&record.ID, &record.Name)
	if err == sql.ErrNoRows {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "DB query failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)
}
