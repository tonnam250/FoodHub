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

type Order struct {
	ID     int    `json:"id"`
	UserID int    `json:"userId"`
	MenuID int    `json:"menuId"`
	Qty    int    `json:"qty"`
	Status string `json:"status"`
}

type UpdateStatus struct {
	Status string `json:"status"`
}

var db *sql.DB

func initDB() {

	connStr := "host=postgres port=5432 user=postgres password=postgres dbname=foodhub sslmode=disable"

	var err error

	db, err = sql.Open("postgres", connStr)

	if err != nil {
		log.Fatal(err) 
	}

	err = db.Ping()

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to PostgreSQL")
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

	resp, err := http.Get(fmt.Sprintf("http://menu-service:3002/menus/%d", menuID))

	if err != nil {
		return false
	}

	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func main() {

	initDB()

	http.HandleFunc("/orders", ordersHandler)
	http.HandleFunc("/orders/", orderByIDHandler)

	log.Println("Order service running on :3003")

	log.Fatal(http.ListenAndServe(":3003", nil))
}

func ordersHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case http.MethodGet:

		rows, err := db.Query("SELECT id,user_id,menu_id,qty,status FROM orders")

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		defer rows.Close()

		var orders []Order

		for rows.Next() {

			var o Order

			rows.Scan(&o.ID, &o.UserID, &o.MenuID, &o.Qty, &o.Status)

			orders = append(orders, o)
		}

		json.NewEncoder(w).Encode(orders)

	case http.MethodPost:

		var newOrder Order

		err := json.NewDecoder(r.Body).Decode(&newOrder)

		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		if !checkUserExists(newOrder.UserID) {
			http.Error(w, "User not found", 400)
			return
		}

		if !checkMenuExists(newOrder.MenuID) {
			http.Error(w, "Menu not found", 400)
			return
		}

		err = db.QueryRow(
			"INSERT INTO orders(user_id,menu_id,qty,status) VALUES($1,$2,$3,$4) RETURNING id",
			newOrder.UserID,
			newOrder.MenuID,
			newOrder.Qty,
			"CREATED",
		).Scan(&newOrder.ID)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		newOrder.Status = "CREATED"

		w.WriteHeader(http.StatusCreated)

		json.NewEncoder(w).Encode(newOrder)

	default:
		http.Error(w, "Method not allowed", 405)
	}
}

func orderByIDHandler(w http.ResponseWriter, r *http.Request) {

	idStr := strings.TrimPrefix(r.URL.Path, "/orders/")

	id, err := strconv.Atoi(idStr)

	if err != nil {
		http.Error(w, "Invalid ID", 400)
		return
	}

	switch r.Method {

	case http.MethodGet:

		var o Order

		err := db.QueryRow(
			"SELECT id,user_id,menu_id,qty,status FROM orders WHERE id=$1",
			id,
		).Scan(&o.ID, &o.UserID, &o.MenuID, &o.Qty, &o.Status)

		if err != nil {
			http.Error(w, "Order not found", 404)
			return
		}

		json.NewEncoder(w).Encode(o)

	case http.MethodPatch:

		var update UpdateStatus

		err := json.NewDecoder(r.Body).Decode(&update)

		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		_, err = db.Exec(
			"UPDATE orders SET status=$1 WHERE id=$2",
			update.Status,
			id,
		)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Write([]byte("Status updated"))

	case http.MethodDelete:

		_, err := db.Exec("DELETE FROM orders WHERE id=$1", id)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Write([]byte("Order deleted"))

	default:
		http.Error(w, "Method not allowed", 405)
	}
}
