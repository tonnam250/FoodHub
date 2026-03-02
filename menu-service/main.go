package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

type Menu struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

var (
	menuStore = make(map[int]Menu)
	nextID    = 1
	mu        sync.Mutex
)

func seedData() {
	menuStore[1] = Menu{ID: 1, Name: "Burger", Price: 99}
	menuStore[2] = Menu{ID: 2, Name: "Pizza", Price: 149}
	nextID = 3
}

// GET /menu
func getMenus(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var list []Menu
	for _, m := range menuStore {
		list = append(list, m)
	}

	json.NewEncoder(w).Encode(list)
}

// GET /menu/{id}
func getMenu(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	menu, exists := menuStore[id]

	if !exists {
		http.Error(w, `{"error":"menu not found"}`, http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(menu)
}

// POST /menu
func createMenu(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var menu Menu
	json.NewDecoder(r.Body).Decode(&menu)

	if menu.Name == "" || menu.Price <= 0 {
		http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
		return
	}

	menu.ID = nextID
	nextID++
	menuStore[menu.ID] = menu

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(menu)
}

// PUT /menu/{id}
func updateMenu(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	_, exists := menuStore[id]
	if !exists {
		http.Error(w, `{"error":"menu not found"}`, http.StatusNotFound)
		return
	}

	var updated Menu
	json.NewDecoder(r.Body).Decode(&updated)

	if updated.Name == "" || updated.Price <= 0 {
		http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
		return
	}

	updated.ID = id
	menuStore[id] = updated

	json.NewEncoder(w).Encode(updated)
}

// DELETE /menu/{id}
func deleteMenu(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	if _, exists := menuStore[id]; !exists {
		http.Error(w, `{"error":"menu not found"}`, http.StatusNotFound)
		return
	}

	delete(menuStore, id)

	json.NewEncoder(w).Encode(map[string]string{
		"message": "menu deleted",
	})
}

func main() {
	seedData()

	r := mux.NewRouter()

	r.HandleFunc("/menu", getMenus).Methods("GET")
	r.HandleFunc("/menu/{id}", getMenu).Methods("GET")
	r.HandleFunc("/menu", createMenu).Methods("POST")
	r.HandleFunc("/menu/{id}", updateMenu).Methods("PUT")
	r.HandleFunc("/menu/{id}", deleteMenu).Methods("DELETE")

	http.ListenAndServe(":3002", r)
}