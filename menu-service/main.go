package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Menu struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

var menus = []Menu{
	{ID: 1, Name: "Salad Bowl", Price: 120},
	{ID: 2, Name: "Protein Smoothie", Price: 95},
}

func main() {
	http.HandleFunc("/menus", menusHandler)
	http.HandleFunc("/menus/", menuByIDHandler)

	log.Println("Menu service running on :3002")
	log.Fatal(http.ListenAndServe(":3002", nil))
}

func menusHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		json.NewEncoder(w).Encode(menus)

	case http.MethodPost:
		var newMenu Menu
		json.NewDecoder(r.Body).Decode(&newMenu)
		newMenu.ID = len(menus) + 1
		menus = append(menus, newMenu)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newMenu)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func menuByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/menus/")
	id, _ := strconv.Atoi(idStr)

	for _, m := range menus {
		if m.ID == id {
			json.NewEncoder(w).Encode(m)
			return
		}
	}

	http.Error(w, "Menu not found", http.StatusNotFound)
}
