package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type CartItem struct {
	ID       int `json:"id"`
	UserID   int `json:"user_id"`
	MenuID   int `json:"menu_id"`
	Quantity int `json:"quantity"`
}

var db *sql.DB

func main() {
	connStr := "postgres://postgres:password@cart-db:5432/cartdb?sslmode=disable"
	var err error

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Database not reachable")
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS cart_items (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL,
		menu_id INT NOT NULL,
		quantity INT NOT NULL
	)
	`)
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/cart", getCart).Methods("GET")
	r.HandleFunc("/cart/{userId}", getCartByUser).Methods("GET")
	r.HandleFunc("/cart", addToCart).Methods("POST")
	r.HandleFunc("/cart/{id}", deleteCartItem).Methods("DELETE")

	log.Println("Cart service running on port 3005")
	log.Fatal(http.ListenAndServe(":3005", enableCORS(r)))
}

//Get
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

func getCart(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, user_id, menu_id, quantity FROM cart_items")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var items []CartItem
	for rows.Next() {
		var item CartItem
		rows.Scan(&item.ID, &item.UserID, &item.MenuID, &item.Quantity)
		items = append(items, item)
	}

	json.NewEncoder(w).Encode(items)
}

// Get cart/(id)
func getCartByUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID, _ := strconv.Atoi(params["userId"])

	rows, err := db.Query("SELECT id, user_id, menu_id, quantity FROM cart_items WHERE user_id=$1", userID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var items []CartItem
	for rows.Next() {
		var item CartItem
		rows.Scan(&item.ID, &item.UserID, &item.MenuID, &item.Quantity)
		items = append(items, item)
	}

	json.NewEncoder(w).Encode(items)
}

// Post
func addToCart(w http.ResponseWriter, r *http.Request) {
	var item CartItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if item.UserID <= 0 || item.MenuID <= 0 || item.Quantity <= 0 {
		http.Error(w, "user_id, menu_id and quantity must be greater than 0", http.StatusBadRequest)
		return
	}

	if !checkMenuExists(item.MenuID) {
		http.Error(w, "menu not found", http.StatusBadRequest)
		return
	}

	var existingID, existingQty int
	err := db.QueryRow(
		"SELECT id, quantity FROM cart_items WHERE user_id=$1 AND menu_id=$2",
		item.UserID, item.MenuID,
	).Scan(&existingID, &existingQty)

	if err == nil {
		item.ID = existingID
		item.Quantity = existingQty + item.Quantity
		_, err = db.Exec("UPDATE cart_items SET quantity=$1 WHERE id=$2", item.Quantity, item.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(item)
		return
	}

	if err != sql.ErrNoRows {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = db.QueryRow(
		"INSERT INTO cart_items (user_id, menu_id, quantity) VALUES ($1,$2,$3) RETURNING id",
		item.UserID, item.MenuID, item.Quantity,
	).Scan(&item.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

// Delete
func deleteCartItem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])

	result, err := db.Exec("DELETE FROM cart_items WHERE id=$1", id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Not found", 404)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

//check before item got add to cart
func checkMenuExists(menuID int) bool {
	url := "http://menu-service:3002/menu/" + strconv.Itoa(menuID)

	resp, err := http.Get(url)
	if err != nil {
		return false
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false
	}

	return true
}
