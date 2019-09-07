package item

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
	purchaseItemRequest struct {
		Item string `json:"item"`
	}

	purchaseItemCompensationRequest struct {
		PurchaseItemID int `json:"purchase_item_id"`
	}

	purchaseItemResponse struct {
		PuchaseItemID int  `json:"purchase_item_id,omitempty"`
		Success       bool `json:"success"`
	}
)

// Serve will serve item service
var Serve = cli.Command{
	Name:        "item",
	Usage:       "Run item service",
	Description: "Execute this command to start item service",
	Action:      startPurchaseItemService,
}

func startPurchaseItemService(c *cli.Context) {
	r := mux.NewRouter()
	r.HandleFunc("/item-success", purchaseItemSuceess).Methods(http.MethodPost)
	r.HandleFunc("/item-failed", purchaseItemFailed).Methods(http.MethodPost)
	r.HandleFunc("/item-compensated", purchaseItemCompensated).Methods(http.MethodPost)

	srv := &http.Server{
		Addr:         ":8001",
		WriteTimeout: time.Second * 3,
		ReadTimeout:  time.Second * 3,
		IdleTimeout:  time.Second * 10,
		Handler:      r,
	}

	go func() {
		log.Println("item service is running on port 8001")
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

func purchaseItemSuceess(w http.ResponseWriter, r *http.Request) {
	var payload purchaseItemRequest
	json.NewDecoder(r.Body).Decode(&payload)

	log.Printf("[purchase item ID 66] purchase item %s : success\n", payload.Item)

	resp := purchaseItemResponse{
		PuchaseItemID: 66,
		Success:       true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func purchaseItemFailed(w http.ResponseWriter, r *http.Request) {
	var payload purchaseItemRequest
	json.NewDecoder(r.Body).Decode(&payload)

	log.Printf("purchase item %s : FAILED!!!\n", payload.Item)

	resp := purchaseItemResponse{
		Success: false,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(resp)
}

func purchaseItemCompensated(w http.ResponseWriter, r *http.Request) {
	var payload purchaseItemCompensationRequest
	json.NewDecoder(r.Body).Decode(&payload)

	log.Printf("[rollback] rollback purchase_item_id %d : success\n", payload.PurchaseItemID)

	resp := purchaseItemResponse{
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
