package main

import (
	"encoding/json"
	"fmt"
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

func checkOrderExists(orderID int) bool {
	resp, err := http.Get(fmt.Sprintf("http://order-service:3003/orders/%d", orderID))
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}
	return true
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
		if err := json.NewDecoder(r.Body).Decode(&newPayment); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if !checkOrderExists(newPayment.OrderID) {
			http.Error(w, "Order not found", http.StatusBadRequest)
			return
		}

		newPayment.ID = len(payments) + 1
		newPayment.Status = "PAID"
		payments = append(payments, newPayment)

		w.Header().Set("Content-Type", "application/json")
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
