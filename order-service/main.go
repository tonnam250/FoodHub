package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Order struct {
	ID     int    `json:"id"`
	UserID int    `json:"userId"`
	MenuID int    `json:"menuId"`
	Qty    int    `json:"qty"`
	Status string `json:"status"`
}

var orders = []Order{
	{ID: 1, UserID: 1, MenuID: 2, Qty: 2, Status: "CREATED"},
}

func checkUserExists(userID int) bool {
	resp, err := http.Get(fmt.Sprintf("http://user-service:3001/users/%d", userID))
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}

func checkMenuExists(menuID int) bool {
	resp, err := http.Get(fmt.Sprintf("http://product-service:3002/menus/%d", menuID))
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}

func main() {
	http.HandleFunc("/orders", ordersHandler)
	http.HandleFunc("/orders/", orderByIDHandler)

	log.Println("Order service running on :3003")
	log.Fatal(http.ListenAndServe(":3003", nil))
}

func ordersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		json.NewEncoder(w).Encode(orders)

	case http.MethodPost:
		var newOrder Order
		if err := json.NewDecoder(r.Body).Decode(&newOrder); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
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

		newOrder.ID = len(orders) + 1
		newOrder.Status = "CREATED"
		orders = append(orders, newOrder)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newOrder)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func orderByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/orders/")
	id, _ := strconv.Atoi(idStr)

	for _, o := range orders {
		if o.ID == id {
			json.NewEncoder(w).Encode(o)
			return
		}
	}

	http.Error(w, "Order not found", http.StatusNotFound)
}
