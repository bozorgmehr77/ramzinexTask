package main

import (
	"encoding/json"
	"fmt"

	"github.com/Shopify/sarama"
)

const (
	KafkaTopic  = "OrderBook"
	KafkaBroker = "localhost:9092"
	timerValue  = 2000
)

type Order struct {
	OrderID int     `json:"order_id"`
	Side    string  `json:"side"`
	Symbol  string  `json:"symbol"`
	Amount  int     `json:"amount"`
	Price   float64 `json:"price"`
}

type OrderBook struct {
	LastUpdateID int         `json:"lastUpdateId"`
	Bids         [][]float64 `json:"bids"`
	Asks         [][]float64 `json:"asks"`
}

func EncodeOrder(order *Order) ([]byte, error) {
	return json.Marshal(order)
}

func DecodeOrder(msg []byte) (Order, error) {
	var order Order
	err := json.Unmarshal(msg, &order)
	if err != nil {
		return Order{}, err
	}
	return order, nil
}

func NewSyncProducer(brokerList []string, config *sarama.Config) (sarama.SyncProducer, error) {
	return sarama.NewSyncProducer(brokerList, config)
}

func NewConsumer(brokerList []string, config *sarama.Config) (sarama.Consumer, error) {
	return sarama.NewConsumer(brokerList, config)
}

func getLastPrice(symbol string) (float64, error) {
	// Define a map of symbol to last price
	lastPrices := map[string]float64{
		"BTCUSDT": 100.0,
		"BTCETH":  200.0,
		"BTCIRT":  500.0,
		// Add more symbols and their last prices as needed
	}

	// Check if the symbol exists in the map
	lastPrice, ok := lastPrices[symbol]
	if !ok {
		return 0, fmt.Errorf("symbol not found")
	}

	return lastPrice, nil
}
