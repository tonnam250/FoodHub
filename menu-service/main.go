package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Menu struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	ImageURL string  `json:"image_url"`
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
		price NUMERIC NOT NULL CHECK (price > 0),
		image_url TEXT
	);`

	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("ALTER TABLE menus ADD COLUMN IF NOT EXISTS image_url TEXT")
	if err != nil {
		log.Fatal(err)
	}
}

func getMenus(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name, price, COALESCE(image_url, '') FROM menus ORDER BY id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	menus := []Menu{}
	for rows.Next() {
		var m Menu
		err := rows.Scan(&m.ID, &m.Name, &m.Price, &m.ImageURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		menus = append(menus, m)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(menus)
}

func getMenu(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	var m Menu
	err = db.QueryRow(
		"SELECT id, name, price, COALESCE(image_url, '') FROM menus WHERE id=$1",
		id,
	).Scan(&m.ID, &m.Name, &m.Price, &m.ImageURL)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"menu not found"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

func createMenu(w http.ResponseWriter, r *http.Request) {
	var m Menu
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
		return
	}

	m.Name = strings.TrimSpace(m.Name)
	m.ImageURL = strings.TrimSpace(m.ImageURL)
	if m.Name == "" || m.Price <= 0 {
		http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
		return
	}

	err := db.QueryRow(
		"INSERT INTO menus (name, price, image_url) VALUES ($1, $2, $3) RETURNING id",
		m.Name, m.Price, m.ImageURL,
	).Scan(&m.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m)
}

func updateMenu(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	var m Menu
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
		return
	}

	m.Name = strings.TrimSpace(m.Name)
	m.ImageURL = strings.TrimSpace(m.ImageURL)
	if m.Name == "" || m.Price <= 0 {
		http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
		return
	}

	result, err := db.Exec(
		"UPDATE menus SET name=$1, price=$2, image_url=$3 WHERE id=$4",
		m.Name, m.Price, m.ImageURL, id,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, `{"error":"menu not found"}`, http.StatusNotFound)
		return
	}

	m.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

func deleteMenu(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	result, err := db.Exec("DELETE FROM menus WHERE id=$1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, `{"error":"menu not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "menu deleted"})
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	if err := db.Ping(); err != nil {
		http.Error(w, "db not ready", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	initDB()

	r := mux.NewRouter()
	r.HandleFunc("/health", healthHandler).Methods("GET")
	r.Handle("/metrics", promhttp.Handler()).Methods("GET")
	r.HandleFunc("/menu", getMenus).Methods("GET")
	r.HandleFunc("/menu/{id}", getMenu).Methods("GET")
	r.HandleFunc("/menu", createMenu).Methods("POST")
	r.HandleFunc("/menu/{id}", updateMenu).Methods("PUT")
	r.HandleFunc("/menu/{id}", deleteMenu).Methods("DELETE")

	log.Println("Menu service running on port 3002")
	log.Fatal(http.ListenAndServe(":3002", enableCORS(r)))
}
