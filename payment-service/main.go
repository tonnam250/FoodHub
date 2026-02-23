package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Payment struct {
	ID      int     `json:"id"`
	OrderID int     `json:"orderId"`
	Amount  float64 `json:"amount"`
	Status  string  `json:"status"`
}

var payments = []Payment{
	{ID: 1, OrderID: 1, Amount: 190, Status: "PAID"},
}

func main() {
	http.HandleFunc("/payments", paymentsHandler)
	http.HandleFunc("/payments/", paymentByIDHandler)

	log.Println("Payment service running on :3004")
	log.Fatal(http.ListenAndServe(":3004", nil))
}

func paymentsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		json.NewEncoder(w).Encode(payments)

	case http.MethodPost:
		var newPayment Payment
		json.NewDecoder(r.Body).Decode(&newPayment)
		newPayment.ID = len(payments) + 1
		newPayment.Status = "PAID"
		payments = append(payments, newPayment)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newPayment)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func paymentByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/payments/")
	id, _ := strconv.Atoi(idStr)

	for _, p := range payments {
		if p.ID == id {
			json.NewEncoder(w).Encode(p)
			return
		}
	}

	http.Error(w, "Payment not found", http.StatusNotFound)
}
