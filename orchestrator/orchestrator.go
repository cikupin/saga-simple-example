package orchestrator

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	saga "github.com/cikupin/go-saga"
	_ "github.com/cikupin/go-saga/storage/kafka" // use kafka as saga log storage engine
	"github.com/gorilla/mux"
	"github.com/urfave/cli"
)

type (
	buyItemRequest struct {
		Item          string `json:"item"`
		Price         int    `json:"price"`
		PaymentMethod string `json:"payment_method"`
	}

	buyItemResponse struct {
		Success bool `json:"success"`
	}
)

var (
	// Serve will serve saga orchestrator
	Serve = cli.Command{
		Name:        "main",
		Usage:       "Run saga orchestrator",
		Description: "Execute this command to start saga orchestrator",
		Action:      startOrchestrator,
	}

	sagaOnce sync.Once
)

const (
	sagaTopicID       = 12
	labelPurchaseItem = "purchase-item"
	labelOrder        = "order"
	labelPayment      = "payment"
)

type orderProperty struct {
	PurchaseItemID int
	OrderID        int
}

func init() {
	sagaOnce.Do(func() {
		saga.StorageConfig.Kafka.ZkAddrs = []string{"0.0.0.0:2181"}
		saga.StorageConfig.Kafka.BrokerAddrs = []string{"0.0.0.0:9092"}
		saga.StorageConfig.Kafka.Partitions = 1
		saga.StorageConfig.Kafka.Replicas = 1
		saga.StorageConfig.Kafka.ReturnDuration = 50 * time.Millisecond
	})
}

func startOrchestrator(c *cli.Context) {
	r := mux.NewRouter()
	r.HandleFunc("/normal-flow", handlerNormalFlow).Methods(http.MethodPost)
	r.HandleFunc("/purchase-failed", handlerPurchaseItemFailed).Methods(http.MethodPost)
	r.HandleFunc("/order-failed", handlerOrderFailed).Methods(http.MethodPost)
	r.HandleFunc("/payment-failed", handlerPaymentFailed).Methods(http.MethodPost)

	srv := &http.Server{
		Addr:         ":8000",
		WriteTimeout: time.Second * 3,
		ReadTimeout:  time.Second * 3,
		IdleTimeout:  time.Second * 10,
		Handler:      r,
	}

	go func() {
		log.Println("order service is running on port 8000")
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

func getInput(r *http.Request) (buyItemRequest, error) {
	var req buyItemRequest
	var err error

	err = json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

// handlerNormalFlow defines normal flow handler
func handlerNormalFlow(w http.ResponseWriter, r *http.Request) {
	input, err := getInput(r)
	if err != nil {
		log.Panicln(err.Error())
		return
	}

	saga.AddSubTxDef(labelPurchaseItem, purchaseItemSuccess, compensatePurchaseItem).
		AddSubTxDef(labelOrder, orderSuccess, compensateOrder).
		AddSubTxDef(labelPayment, paymentSuccess, compensatePayment)

	ctx := context.Background()
	property := &orderProperty{}
	sagaInstance := saga.StartSaga(ctx, sagaTopicID).
		ExecSub(labelPurchaseItem, property, input.Item).
		ExecSub(labelOrder, property, input.Item, input.Price).
		ExecSub(labelPayment, property, input.PaymentMethod, input.Price).
		EndSaga()

	generateResponse(w, sagaInstance.IsAborted())
	return
}

// handlerPurchaseItemFailed defines puchase item failed hanlder
func handlerPurchaseItemFailed(w http.ResponseWriter, r *http.Request) {
	input, err := getInput(r)
	if err != nil {
		log.Panicln(err.Error())
		return
	}

	saga.AddSubTxDef(labelPurchaseItem, purchaseItemFailed, compensatePurchaseItem).
		AddSubTxDef(labelOrder, orderSuccess, compensateOrder).
		AddSubTxDef(labelPayment, paymentSuccess, compensatePayment)

	ctx := context.Background()
	property := &orderProperty{}
	sagaInstance := saga.StartSaga(ctx, sagaTopicID).
		ExecSub(labelPurchaseItem, property, input.Item).
		ExecSub(labelOrder, property, input.Item, input.Price).
		ExecSub(labelPayment, property, input.PaymentMethod, input.Price).
		EndSaga()

	generateResponse(w, sagaInstance.IsAborted())
	return
}

// handlerOrderFailed defines order failed handler
func handlerOrderFailed(w http.ResponseWriter, r *http.Request) {
	input, err := getInput(r)
	if err != nil {
		log.Panicln(err.Error())
		return
	}

	saga.AddSubTxDef(labelPurchaseItem, purchaseItemSuccess, compensatePurchaseItem).
		AddSubTxDef(labelOrder, orderFailed, compensateOrder).
		AddSubTxDef(labelPayment, paymentSuccess, compensatePayment)

	ctx := context.Background()
	property := &orderProperty{}
	sagaInstance := saga.StartSaga(ctx, sagaTopicID).
		ExecSub(labelPurchaseItem, property, input.Item).
		ExecSub(labelOrder, property, input.Item, input.Price).
		ExecSub(labelPayment, property, input.PaymentMethod, input.Price).
		EndSaga()

	generateResponse(w, sagaInstance.IsAborted())
	return
}

// handlerPaymentFailed defines payment failed handler
func handlerPaymentFailed(w http.ResponseWriter, r *http.Request) {
	input, err := getInput(r)
	if err != nil {
		log.Panicln(err.Error())
		return
	}

	saga.AddSubTxDef(labelPurchaseItem, purchaseItemSuccess, compensatePurchaseItem).
		AddSubTxDef(labelOrder, orderSuccess, compensateOrder).
		AddSubTxDef(labelPayment, paymentFailed, compensatePayment)

	ctx := context.Background()
	property := &orderProperty{}
	sagaInstance := saga.StartSaga(ctx, sagaTopicID).
		ExecSub(labelPurchaseItem, property, input.Item).
		ExecSub(labelOrder, property, input.Item, input.Price).
		ExecSub(labelPayment, property, input.PaymentMethod, input.Price).
		EndSaga()

	generateResponse(w, sagaInstance.IsAborted())
	return
}

func generateResponse(w http.ResponseWriter, isAborted bool) {
	response := buyItemResponse{Success: !isAborted}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
