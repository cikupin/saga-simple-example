package orchestrator

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/lysu/go-saga"
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

// Serve will serve saga orchestrator
var Serve = cli.Command{
	Name:        "main",
	Usage:       "Run saga orchestrator",
	Description: "Execute this command to start saga orchestrator",
	Action:      startOrchestrator,
}

var (
	contextKeyPurchaseItemID = "purchase-item-id"
	contextKeyOrderID        = "order-id"
)

const (
	sagaTopicID       = 12
	labelPurchaseItem = "purchase-item"
	labelOrder        = "order"
	labelPayment      = "payment"
)

func init() {

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

	saga.StorageConfig.Kafka.ZkAddrs = []string{"0.0.0.0:2181"}
	saga.StorageConfig.Kafka.BrokerAddrs = []string{"0.0.0.0:9092"}
	saga.StorageConfig.Kafka.Partitions = 1
	saga.StorageConfig.Kafka.Replicas = 1
	saga.StorageConfig.Kafka.ReturnDuration = 50 * time.Millisecond

	saga.AddSubTxDef(labelPurchaseItem, purchaseItemSuccess, purchaseItemCompensation).
		AddSubTxDef(labelOrder, orderSuccess, orderCompensation).
		AddSubTxDef(labelPayment, paymentSuccess, paymentFailed)

	ctx := context.Background()
	saga.StartSaga(ctx, sagaTopicID).
		ExecSub(labelPurchaseItem, input.Item).
		ExecSub(labelOrder, input.Item, input.Price).
		ExecSub(labelPayment, input.PaymentMethod, input.Price, ctx.Value(contextKeyOrderID).(int)).
		EndSaga()
	return
}

// handlerPurchaseItemFailed defines puchase item failed hanlder
func handlerPurchaseItemFailed(w http.ResponseWriter, r *http.Request) {
	input, err := getInput(r)
	if err != nil {
		log.Panicln(err.Error())
		return
	}

	saga.AddSubTxDef(labelPurchaseItem, purchaseItemFailed, purchaseItemCompensation).
		AddSubTxDef(labelOrder, orderSuccess, orderCompensation).
		AddSubTxDef(labelPayment, paymentSuccess, paymentFailed)

	ctx := context.Background()
	saga.StartSaga(ctx, sagaTopicID).
		ExecSub(labelPurchaseItem, input.Item).
		ExecSub(labelOrder, input.Item, input.Price).
		ExecSub(labelPayment, input.PaymentMethod, input.Price, ctx.Value(contextKeyOrderID).(int))
	return
}

// handlerOrderFailed defines order failed handler
func handlerOrderFailed(w http.ResponseWriter, r *http.Request) {
	input, err := getInput(r)
	if err != nil {
		log.Panicln(err.Error())
		return
	}

	saga.AddSubTxDef(labelPurchaseItem, purchaseItemSuccess, purchaseItemCompensation).
		AddSubTxDef(labelOrder, orderFailed, orderCompensation).
		AddSubTxDef(labelPayment, paymentSuccess, paymentFailed)

	ctx := context.Background()
	saga.StartSaga(ctx, sagaTopicID).
		ExecSub(labelPurchaseItem, input.Item).
		ExecSub(labelOrder, input.Item, input.Price).
		ExecSub(labelPayment, input.PaymentMethod, input.Price, ctx.Value(contextKeyOrderID).(int))
	return
}

// handlerPaymentFailed defines payment failed handler
func handlerPaymentFailed(w http.ResponseWriter, r *http.Request) {
	input, err := getInput(r)
	if err != nil {
		log.Panicln(err.Error())
		return
	}

	saga.AddSubTxDef(labelPurchaseItem, purchaseItemSuccess, purchaseItemCompensation).
		AddSubTxDef(labelOrder, orderSuccess, orderCompensation).
		AddSubTxDef(labelPayment, paymentFailed, paymentFailed)

	ctx := context.Background()
	saga.StartSaga(ctx, sagaTopicID).
		ExecSub(labelPurchaseItem, input.Item).
		ExecSub(labelOrder, input.Item, input.Price).
		ExecSub(labelPayment, input.PaymentMethod, input.Price, ctx.Value(contextKeyOrderID).(int))
	return
}
