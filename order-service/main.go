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

type UpdateStatus struct {
	Status string `json:"status"`
}

var orders = []Order{
	{ID: 1, UserID: 1, MenuID: 2, Qty: 2, Status: "CREATED"},
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
	http.HandleFunc("/orders", ordersHandler)
	http.HandleFunc("/orders/", orderByIDHandler)

	log.Println("Order service running on :3003")
	log.Fatal(http.ListenAndServe(":3003", nil))
}

func ordersHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case http.MethodGet:

		w.Header().Set("Content-Type", "application/json")
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
	id, err := strconv.Atoi(idStr)

	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	switch r.Method {

	case http.MethodGet:

		for _, o := range orders {
			if o.ID == id {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(o)
				return
			}
		}

		http.Error(w, "Order not found", http.StatusNotFound)

	case http.MethodDelete:

		for i, o := range orders {
			if o.ID == id {

				orders = append(orders[:i], orders[i+1:]...)

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{
					"message": "Order cancelled",
				})
				return
			}
		}

		http.Error(w, "Order not found", http.StatusNotFound)

	case http.MethodPatch:

		var update UpdateStatus

		err := json.NewDecoder(r.Body).Decode(&update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for i, o := range orders {
			if o.ID == id {

				orders[i].Status = update.Status

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(orders[i])
				return
			}
		}

		http.Error(w, "Order not found", http.StatusNotFound)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
