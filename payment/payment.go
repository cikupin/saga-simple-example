package payment

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/urfave/cli"
)

type (
	paymentRequest struct {
		PaymentMethod string `json:"payment_method"`
		Amount        int    `json:"amount"`
		OrderID       int    `json:"order_id"`
	}

	paymentResponse struct {
		PaymentID int  `json:"payment_id,omitempty"`
		Success   bool `json:"success"`
	}
)

// Serve will serve payment service
var Serve = cli.Command{
	Name:        "payment",
	Usage:       "Run payment service",
	Description: "Execute this command to start payment service",
	Action:      startPaymentService,
}

// startPaymentService wil start payment service
func startPaymentService(c *cli.Context) {
	r := mux.NewRouter()
	r.HandleFunc("/payment-success", paymentSucess).Methods(http.MethodPost)
	r.HandleFunc("/payment-failed", paymentFailed).Methods(http.MethodPost)

	srv := &http.Server{
		Addr:         ":8003",
		WriteTimeout: time.Second * 3,
		ReadTimeout:  time.Second * 3,
		IdleTimeout:  time.Second * 10,
		Handler:      r,
	}

	go func() {
		log.Println("payment service is running on port 8003")
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	chanSignal := make(chan os.Signal, 1)
	signal.Notify(chanSignal, os.Interrupt)
	<-chanSignal

	// 3 seconds graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	srv.Shutdown(ctx)
	log.Println("shutting down")
	os.Exit(0)
}

// paymentSucess defines payment success logic
func paymentSucess(w http.ResponseWriter, r *http.Request) {
	var payload paymentRequest
	json.NewDecoder(r.Body).Decode(&payload)

	log.Printf("[payment ID 10] $%d payment for order_id %d with payment method %s : success\n", payload.Amount, payload.OrderID, payload.PaymentMethod)

	resp := paymentResponse{
		PaymentID: 10,
		Success:   true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// defines payment failed logic
func paymentFailed(w http.ResponseWriter, r *http.Request) {
	var payload paymentRequest
	json.NewDecoder(r.Body).Decode(&payload)

	log.Printf("$%d payment for order_id %d with payment method %s : FAILED!!!\n", payload.Amount, payload.OrderID, payload.PaymentMethod)

	resp := paymentResponse{
		Success: false,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(resp)
}
