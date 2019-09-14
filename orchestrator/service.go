package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/cikupin/saga-simple-example/item"
	"github.com/cikupin/saga-simple-example/order"
	"github.com/cikupin/saga-simple-example/payment"
)

// purchaseItemSuccess will purchase item and success
func purchaseItemSuccess(ctx context.Context, itemName string) error {
	payload := item.Request{
		Item: itemName,
	}
	payloadBytes, _ := json.Marshal(payload)

	resp, err := http.Post("http://localhost:8001/item-success", "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		log.Panicln(err.Error())
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panicln(err.Error())
		return err
	}

	var response item.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Panicln(err.Error())
		return err
	}

	ctx = context.WithValue(ctx, contextKeyPurchaseItemID, response.PuchaseItemID)
	return nil
}

// purchaseItemFailed will purchase item and failed
func purchaseItemFailed(ctx context.Context, itemName string) error {
	// payload := item.Request{
	// 	Item: itemName,
	// }
	return nil
}

// purchaseItemCompensation will rollback purchase item
func purchaseItemCompensation(ctx context.Context, itemName string) error {
	// payload := item.Request{
	// 	Item: itemName,
	// }
	return nil
}

// orderSuccess will record order data and success
func orderSuccess(ctx context.Context, itemName string, price int) error {
	payload := order.Request{
		Item:  itemName,
		Price: price,
	}
	payloadBytes, _ := json.Marshal(payload)

	resp, err := http.Post("http://localhost:8001/order-success", "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		log.Panicln(err.Error())
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panicln(err.Error())
		return err
	}

	var response order.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Panicln(err.Error())
		return err
	}

	ctx = context.WithValue(ctx, contextKeyOrderID, response.OrderID)
	return nil
}

// orderFailed will record order data and failed
func orderFailed(ctx context.Context, itemName string, price int) error {
	return nil
}

// orderCompensation will rollback order
func orderCompensation(ctx context.Context, itemName string, price int) error {
	return nil
}

// paymentSuccess will do payment and success
func paymentSuccess(ctx context.Context, paymentMethod string, price, orderID int) error {
	payload := payment.Request{
		PaymentMethod: paymentMethod,
		Price:         price,
		OrderID:       orderID,
	}
	payloadBytes, _ := json.Marshal(payload)

	resp, err := http.Post("http://localhost:8001/payment-success", "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		log.Panicln(err.Error())
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panicln(err.Error())
		return err
	}

	var response payment.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Panicln(err.Error())
		return err
	}
	return nil
}

// paymentFailed will do payment and failed
func paymentFailed(ctx context.Context, paymentMethod string, price, orderID int) error {
	return nil
}
