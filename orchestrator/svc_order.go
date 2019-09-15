package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/cikupin/saga-simple-example/item"
	"github.com/cikupin/saga-simple-example/order"
)

// orderSuccess will record order data and success
func orderSuccess(ctx context.Context, prop *orderProperty, itemName string, price int) error {
	payload := order.Request{
		Item:  itemName,
		Price: price,
	}
	payloadBytes, _ := json.Marshal(payload)

	resp, err := http.Post("http://localhost:8002/order-success", "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		log.Println(err.Error())
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 500 {
		err = errors.New("request error to service order")
		log.Println(err.Error())
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	var response order.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	prop.OrderID = response.OrderID
	return nil
}

// orderFailed will record order data and failed
func orderFailed(ctx context.Context, prop *orderProperty, itemName string, price int) error {
	payload := order.Request{
		Item:  itemName,
		Price: price,
	}
	payloadBytes, _ := json.Marshal(payload)

	resp, err := http.Post("http://localhost:8002/order-failed", "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		log.Println(err.Error())
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 500 {
		err = errors.New("request error to service order")
		log.Println(err.Error())
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	var response order.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	prop.OrderID = response.OrderID
	return nil
}

// compensateOrder will rollback item purchase if fail to request order
func compensateOrder(ctx context.Context, prop *orderProperty, itemName string, price int) error {
	payload := item.CompensationRequest{
		PurchaseItemID: prop.PurchaseItemID,
	}
	payloadBytes, _ := json.Marshal(payload)

	resp, err := http.Post("http://localhost:8001/item-compensated", "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		log.Println(err.Error())
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 500 {
		err = errors.New("request error to rollback item")
		log.Println(err.Error())
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	var response item.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	prop.PurchaseItemID = response.PuchaseItemID
	return nil
}
