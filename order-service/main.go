package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Order struct {
	ID     int    `json:"id"`
	UserID int    `json:"userId"`
	MenuID int    `json:"menuId"`
	Qty    int    `json:"qty"`
	Status string `json:"status"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
}

var db *sql.DB

func main() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:password@order-db:5432/orderdb?sslmode=disable"
	}

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Database not reachable")
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS orders (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL,
		menu_id INT NOT NULL,
		qty INT NOT NULL CHECK (qty > 0),
		status TEXT NOT NULL
	)
	`)
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/orders", getOrders).Methods("GET")
	r.HandleFunc("/orders/{id}", getOrderByID).Methods("GET")
	r.HandleFunc("/orders", createOrder).Methods("POST")
	r.HandleFunc("/orders/{id}/status", updateOrderStatus).Methods("PUT")

	log.Println("Order service running on :3003")
	log.Fatal(http.ListenAndServe(":3003", enableCORS(r)))
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

func getOrders(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, user_id, menu_id, qty, status FROM orders ORDER BY id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	orders := []Order{}
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.MenuID, &o.Qty, &o.Status); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		orders = append(orders, o)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func getOrderByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var o Order
	err = db.QueryRow(
		"SELECT id, user_id, menu_id, qty, status FROM orders WHERE id=$1",
		id,
	).Scan(&o.ID, &o.UserID, &o.MenuID, &o.Qty, &o.Status)

	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(o)
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	var newOrder Order
	if err := json.NewDecoder(r.Body).Decode(&newOrder); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if newOrder.Qty <= 0 {
		http.Error(w, "qty must be greater than 0", http.StatusBadRequest)
		return
	}

	if !checkUserExists(newOrder.UserID) {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	if !checkMenuExists(newOrder.MenuID) {
		http.Error(w, "Menu not found", http.StatusBadRequest)
		return
	}

	newOrder.Status = "CREATED"
	if err := db.QueryRow(
		"INSERT INTO orders (user_id, menu_id, qty, status) VALUES ($1,$2,$3,$4) RETURNING id",
		newOrder.UserID, newOrder.MenuID, newOrder.Qty, newOrder.Status,
	).Scan(&newOrder.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newOrder)
}

func updateOrderStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	target := normalizeStatus(req.Status)
	if target == "" {
		http.Error(w, "status is required", http.StatusBadRequest)
		return
	}

	var order Order
	err = db.QueryRow(
		"SELECT id, user_id, menu_id, qty, status FROM orders WHERE id=$1",
		id,
	).Scan(&order.ID, &order.UserID, &order.MenuID, &order.Qty, &order.Status)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	current := normalizeStatus(order.Status)
	if !isValidTransition(current, target) {
		http.Error(w, "invalid status transition", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE orders SET status=$1 WHERE id=$2", target, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	order.Status = target
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func checkUserExists(userID int) bool {
	resp, err := http.Get(fmt.Sprintf("http://user-service:3001/users/%d", userID))
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func checkMenuExists(menuID int) bool {
	resp, err := http.Get(fmt.Sprintf("http://menu-service:3002/menu/%d", menuID))
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func normalizeStatus(status string) string {
	switch status {
	case "CREATED", "PREPARING", "READY", "PICKED_UP":
		return status
	default:
		return ""
	}
}

func isValidTransition(current, target string) bool {
	if current == target {
		return true
	}
	allowed := map[string]string{
		"CREATED":   "PREPARING",
		"PREPARING": "READY",
		"READY":     "PICKED_UP",
	}
	return allowed[current] == target
}
