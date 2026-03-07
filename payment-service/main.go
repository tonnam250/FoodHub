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
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

type Payment struct {
	ID        int     `json:"id"`
	OrderID   int     `json:"orderId"`
	Amount    float64 `json:"amount"`
	Method    string  `json:"method"`
	Status    string  `json:"status"`
	Reference string  `json:"reference"`
	PaidAt    string  `json:"paidAt,omitempty"`
}

type CreatePaymentRequest struct {
	OrderID      int     `json:"orderId"`
	Amount       float64 `json:"amount"`
	Method       string  `json:"method"`
	SimulateFail bool    `json:"simulateFail"`
}

var db *sql.DB

func main() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:password@payment-db:5432/paymentdb?sslmode=disable"
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
	CREATE TABLE IF NOT EXISTS payments (
		id SERIAL PRIMARY KEY,
		order_id INT NOT NULL UNIQUE,
		amount NUMERIC(10,2) NOT NULL CHECK (amount > 0),
		method TEXT NOT NULL DEFAULT 'cash',
		status TEXT NOT NULL,
		reference TEXT NOT NULL,
		paid_at TIMESTAMP NULL
	)
	`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("ALTER TABLE payments ADD COLUMN IF NOT EXISTS method TEXT NOT NULL DEFAULT 'cash'")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("ALTER TABLE payments ADD COLUMN IF NOT EXISTS reference TEXT NOT NULL DEFAULT ''")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("ALTER TABLE payments ADD COLUMN IF NOT EXISTS paid_at TIMESTAMP NULL")
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/payments", getPayments).Methods("GET")
	r.HandleFunc("/payments/{id}", getPaymentByID).Methods("GET")
	r.HandleFunc("/payments", createPayment).Methods("POST")

	log.Println("Payment service running on :3004")
	log.Fatal(http.ListenAndServe(":3004", enableCORS(r)))
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

func getPayments(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, order_id, amount, method, status, reference, paid_at FROM payments ORDER BY id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	payments := []Payment{}
	for rows.Next() {
		var p Payment
		var paidAt sql.NullTime
		if err := rows.Scan(&p.ID, &p.OrderID, &p.Amount, &p.Method, &p.Status, &p.Reference, &paidAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if paidAt.Valid {
			p.PaidAt = paidAt.Time.Format(time.RFC3339)
		}
		payments = append(payments, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payments)
}

func getPaymentByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var p Payment
	var paidAt sql.NullTime
	err = db.QueryRow(
		"SELECT id, order_id, amount, method, status, reference, paid_at FROM payments WHERE id=$1",
		id,
	).Scan(&p.ID, &p.OrderID, &p.Amount, &p.Method, &p.Status, &p.Reference, &paidAt)

	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Payment not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if paidAt.Valid {
		p.PaidAt = paidAt.Time.Format(time.RFC3339)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func createPayment(w http.ResponseWriter, r *http.Request) {
	var req CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		http.Error(w, "amount must be greater than 0", http.StatusBadRequest)
		return
	}

	if !checkOrderExists(req.OrderID) {
		http.Error(w, "Order not found", http.StatusBadRequest)
		return
	}

	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = "CASH"
	}
	if method != "CASH" && method != "QR" && method != "CARD" {
		http.Error(w, "invalid payment method", http.StatusBadRequest)
		return
	}

	status := "PAID"
	if req.SimulateFail {
		status = "FAILED"
	}

	reference := fmt.Sprintf("PMT-%d", time.Now().UnixNano())
	var paymentID int
	var paidAt sql.NullTime
	if status == "PAID" {
		paidAt = sql.NullTime{Time: time.Now(), Valid: true}
	}

	err := db.QueryRow(
		"INSERT INTO payments (order_id, amount, method, status, reference, paid_at) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id",
		req.OrderID, req.Amount, method, status, reference, paidAt,
	).Scan(&paymentID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			http.Error(w, "Order already has payment record", http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	payment := Payment{
		ID:        paymentID,
		OrderID:   req.OrderID,
		Amount:    req.Amount,
		Method:    method,
		Status:    status,
		Reference: reference,
	}
	if paidAt.Valid {
		payment.PaidAt = paidAt.Time.Format(time.RFC3339)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(payment)
}

func checkOrderExists(orderID int) bool {
	resp, err := http.Get(fmt.Sprintf("http://order-service:3003/orders/%d", orderID))
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
