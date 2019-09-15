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
)

// purchaseItemSuccess will purchase item and success
func purchaseItemSuccess(ctx context.Context, prop *orderProperty, itemName string) error {
	payload := item.Request{
		Item: itemName,
	}
	payloadBytes, _ := json.Marshal(payload)

	resp, err := http.Post("http://localhost:8001/item-success", "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		log.Println(err.Error())
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 500 {
		err = errors.New("request error to service item")
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

// purchaseItemFailed will purchase item and failed
func purchaseItemFailed(ctx context.Context, prop *orderProperty, itemName string) error {
	payload := item.Request{
		Item: itemName,
	}
	payloadBytes, _ := json.Marshal(payload)

	resp, err := http.Post("http://localhost:8001/item-failed", "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		log.Println(err.Error())
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 500 {
		err = errors.New("request error to service item")
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

// compensatePurchaseItem will return nil
func compensatePurchaseItem(ctx context.Context, prop *orderProperty, itemName string) error {
	return nil
}
