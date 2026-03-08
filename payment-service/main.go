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

	"FoodHub/common/resilience"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	amqp "github.com/rabbitmq/amqp091-go"
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

type OrderCreatedEvent struct {
	OrderID int     `json:"orderId"`
	UserID  int     `json:"userId"`
	MenuID  int     `json:"menuId"`
	Qty     int     `json:"qty"`
	Amount  float64 `json:"amount"`
	Method  string  `json:"method"`
}

var db *sql.DB
var orderCB = resilience.NewCircuitBreaker(3, 10*time.Second)
var downstreamHTTP = &http.Client{Timeout: 3 * time.Second}

var rabbitConn *amqp.Connection
var rabbitCh *amqp.Channel

const orderCreatedQueue = "order.created"

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

	initRabbitMQ()
	startOrderCreatedConsumer()

	r := mux.NewRouter()
	r.HandleFunc("/health", healthHandler).Methods("GET")
	r.Handle("/metrics", promhttp.Handler()).Methods("GET")
	r.HandleFunc("/payments", getPayments).Methods("GET")
	r.HandleFunc("/payments/{id}", getPaymentByID).Methods("GET")
	r.HandleFunc("/payments", createPayment).Methods("POST")

	log.Println("Payment service running on :3004")
	log.Fatal(http.ListenAndServe(":3004", enableCORS(r)))
}

func initRabbitMQ() {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@rabbitmq:5672/"
	}

	var err error
	for i := 0; i < 20; i++ {
		rabbitConn, err = amqp.Dial(rabbitURL)
		if err == nil {
			break
		}
		log.Printf("rabbitmq not ready, retrying: %v", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Printf("rabbitmq disabled: %v", err)
		return
	}

	rabbitCh, err = rabbitConn.Channel()
	if err != nil {
		log.Printf("rabbitmq channel error: %v", err)
		return
	}

	_, err = rabbitCh.QueueDeclare(orderCreatedQueue, true, false, false, false, nil)
	if err != nil {
		log.Printf("rabbitmq queue declare error: %v", err)
	}
}

func startOrderCreatedConsumer() {
	if rabbitCh == nil {
		return
	}
	msgs, err := rabbitCh.Consume(orderCreatedQueue, "payment-service", false, false, false, false, nil)
	if err != nil {
		log.Printf("rabbit consume error: %v", err)
		return
	}

	go func() {
		for msg := range msgs {
			var event OrderCreatedEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				_ = msg.Nack(false, false)
				continue
			}

			method := event.Method
			if method == "" {
				method = "CASH"
			}
			if _, err := createPaymentRecord(CreatePaymentRequest{
				OrderID: event.OrderID,
				Amount:  event.Amount,
				Method:  method,
			}); err != nil {
				log.Printf("async payment create failed for order %d: %v", event.OrderID, err)
				_ = msg.Ack(false)
				continue
			}
			_ = msg.Ack(false)
		}
	}()
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	if err := db.Ping(); err != nil {
		http.Error(w, "db not ready", http.StatusServiceUnavailable)
		return
	}
	status := "ok"
	if rabbitCh == nil {
		status = "degraded"
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": status, "service": "payment-service"})
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
	_ = json.NewEncoder(w).Encode(payments)
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
	_ = json.NewEncoder(w).Encode(p)
}

func createPayment(w http.ResponseWriter, r *http.Request) {
	var req CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	payment, err := createPaymentRecord(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(payment)
}

func createPaymentRecord(req CreatePaymentRequest) (*Payment, error) {
	if req.Amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	if !checkOrderExists(req.OrderID) {
		return nil, errors.New("order not found")
	}

	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = "CASH"
	}
	if method != "CASH" && method != "QR" && method != "CARD" {
		return nil, errors.New("invalid payment method")
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
			// Idempotency for async flow.
			var existing Payment
			var existingPaidAt sql.NullTime
			qErr := db.QueryRow(
				"SELECT id, order_id, amount, method, status, reference, paid_at FROM payments WHERE order_id=$1",
				req.OrderID,
			).Scan(&existing.ID, &existing.OrderID, &existing.Amount, &existing.Method, &existing.Status, &existing.Reference, &existingPaidAt)
			if qErr == nil {
				if existingPaidAt.Valid {
					existing.PaidAt = existingPaidAt.Time.Format(time.RFC3339)
				}
				return &existing, nil
			}
			return nil, errors.New("order already has payment record")
		}
		return nil, err
	}

	payment := &Payment{
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

	return payment, nil
}

func checkOrderExists(orderID int) bool {
	url := fmt.Sprintf("http://order-service:3003/orders/%d", orderID)
	ok := false
	err := orderCB.Execute(func() error {
		resp, err := downstreamHTTP.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			ok = true
			return nil
		}
		return errors.New("order not found")
	})
	return err == nil && ok
}
