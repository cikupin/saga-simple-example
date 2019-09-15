package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/cikupin/saga-simple-example/order"
	"github.com/cikupin/saga-simple-example/payment"
)

// paymentSuccess will do payment and success
func paymentSuccess(ctx context.Context, prop *orderProperty, paymentMethod string, price int) error {
	payload := payment.Request{
		PaymentMethod: paymentMethod,
		Price:         price,
		OrderID:       prop.OrderID,
	}
	payloadBytes, _ := json.Marshal(payload)

	resp, err := http.Post("http://localhost:8003/payment-success", "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		log.Println(err.Error())
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 500 {
		err = errors.New("request error to service payment")
		log.Println(err.Error())
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	var response payment.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

// paymentFailed will do payment and failed
func paymentFailed(ctx context.Context, prop *orderProperty, paymentMethod string, price int) error {
	payload := payment.Request{
		PaymentMethod: paymentMethod,
		Price:         price,
		OrderID:       prop.OrderID,
	}
	payloadBytes, _ := json.Marshal(payload)

	resp, err := http.Post("http://localhost:8003/payment-failed", "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		log.Println(err.Error())
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 500 {
		err = errors.New("request error to service payment")
		log.Println(err.Error())
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	var response payment.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

// compensatePayment will rollback order if fail to request payment
func compensatePayment(ctx context.Context, prop *orderProperty, paymentMethod string, price int) error {
	payload := order.CompensationRequest{
		OrderID: prop.OrderID,
	}
	payloadBytes, _ := json.Marshal(payload)

	resp, err := http.Post("http://localhost:8002/order-compensated", "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		log.Println(err.Error())
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 500 {
		err = errors.New("request error to rollback order")
		log.Println(err.Error())
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	var response payment.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}
