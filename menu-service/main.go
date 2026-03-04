package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Menu struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

var db *sql.DB

func initDB() {
	var err error

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:password@menu-db:5432/menudb?sslmode=disable"
	}

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Database not reachable")
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS menus (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		price NUMERIC NOT NULL
	);`

	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to PostgreSQL")
}

// GET /menu
func getMenus(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name, price FROM menus")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var menus []Menu
	for rows.Next() {
		var m Menu
		rows.Scan(&m.ID, &m.Name, &m.Price)
		menus = append(menus, m)
	}

	json.NewEncoder(w).Encode(menus)
}

// GET /menu/{id}
func getMenu(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	var m Menu
	err := db.QueryRow("SELECT id, name, price FROM menus WHERE id=$1", id).
		Scan(&m.ID, &m.Name, &m.Price)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"menu not found"}`, 404)
		return
	}

	json.NewEncoder(w).Encode(m)
}

// POST /menu
func createMenu(w http.ResponseWriter, r *http.Request) {
	var m Menu
	json.NewDecoder(r.Body).Decode(&m)

	if m.Name == "" || m.Price <= 0 {
		http.Error(w, `{"error":"invalid input"}`, 400)
		return
	}

	err := db.QueryRow(
		"INSERT INTO menus (name, price) VALUES ($1, $2) RETURNING id",
		m.Name, m.Price,
	).Scan(&m.ID)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m)
}

// PUT /menu/{id}
func updateMenu(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	var m Menu
	json.NewDecoder(r.Body).Decode(&m)

	if m.Name == "" || m.Price <= 0 {
		http.Error(w, `{"error":"invalid input"}`, 400)
		return
	}

	result, err := db.Exec(
		"UPDATE menus SET name=$1, price=$2 WHERE id=$3",
		m.Name, m.Price, id,
	)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, `{"error":"menu not found"}`, 404)
		return
	}

	m.ID = id
	json.NewEncoder(w).Encode(m)
}

// DELETE /menu/{id}
func deleteMenu(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	result, err := db.Exec("DELETE FROM menus WHERE id=$1", id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, `{"error":"menu not found"}`, 404)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "menu deleted",
	})
}

func main() {
	initDB()

	r := mux.NewRouter()

	r.HandleFunc("/menu", getMenus).Methods("GET")
	r.HandleFunc("/menu/{id}", getMenu).Methods("GET")
	r.HandleFunc("/menu", createMenu).Methods("POST")
	r.HandleFunc("/menu/{id}", updateMenu).Methods("PUT")
	r.HandleFunc("/menu/{id}", deleteMenu).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":3002", r))
}
