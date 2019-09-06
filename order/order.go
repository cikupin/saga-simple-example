package order

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
	orderRequest struct {
		Item   string `json:"item"`
		Amount int    `json:"amount"`
	}

	orderCompensationRequest struct {
		OrderID int `json:"order_id"`
	}

	orderResponse struct {
		OrderID int  `json:"order_id,omitempty"`
		Success bool `json:"success"`
	}
)

// Serve will serve order service
var Serve = cli.Command{
	Name:        "order",
	Usage:       "Run order service",
	Description: "Execute this command to start order service",
	Action:      startOrderService,
}

// startOrderService will start order service
func startOrderService(c *cli.Context) {
	r := mux.NewRouter()
	r.HandleFunc("/order-success", orderSuccess).Methods(http.MethodPost)
	r.HandleFunc("/order-failed", orderFailed).Methods(http.MethodPost)
	r.HandleFunc("/order-compensated", orderCompensation).Methods(http.MethodPost)

	srv := &http.Server{
		Addr:         ":8002",
		WriteTimeout: time.Second * 3,
		ReadTimeout:  time.Second * 3,
		IdleTimeout:  time.Second * 10,
		Handler:      r,
	}

	go func() {
		log.Println("order service is running on port 8002")
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

// orderSuccess defines order success logic
func orderSuccess(w http.ResponseWriter, r *http.Request) {
	var payload orderRequest
	json.NewDecoder(r.Body).Decode(&payload)

	log.Printf("[order ID 32] purchase item %s for amount $%d : success\n", payload.Item, payload.Amount)

	resp := orderResponse{
		OrderID: 32,
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// orderFailed defines order failed logic
func orderFailed(w http.ResponseWriter, r *http.Request) {
	var payload orderRequest
	json.NewDecoder(r.Body).Decode(&payload)

	log.Printf("purchase item %s for amount $%d : FAILED!!!\n", payload.Item, payload.Amount)

	resp := orderResponse{
		Success: false,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(resp)
}

// orderCompensation defines order compensation logic
func orderCompensation(w http.ResponseWriter, r *http.Request) {
	var payload orderCompensationRequest
	json.NewDecoder(r.Body).Decode(&payload)

	log.Printf("[rollback] rollback order_id %d : success\n", payload.OrderID)

	resp := orderResponse{
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
