package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/Shopify/sarama"
)

func produceSamplesInLoop() {
	for {
		produceSamples()
		time.Sleep(20 * time.Second)
	}
}

func produceSamples() {
	// Set up the Kafka producer configuration
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true

	// Create a new Kafka sync producer
	producer, err := sarama.NewSyncProducer([]string{KafkaBroker}, config)
	if err != nil {
		log.Fatalln("Failed to start Kafka producer:", err)
	}
	defer producer.Close()

	// Generate ten sample orders and send them to Kafka

	lastPrice1 := 100.0
	symbol1 := "BTCUSDT"
	orders := generateOrders(lastPrice1, symbol1)
	lastPrice2 := 200.0
	symbol2 := "BTCETH"
	orders2 := generateOrders(lastPrice2, symbol2)
	lastPrice3 := 500.0
	symbol3 := "BTCIRT"
	orders3 := generateOrders(lastPrice3, symbol3)
	orders = append(orders, orders2...)
	orders = append(orders, orders3...)

	for _, order := range orders {
		orderBytes, err := json.Marshal(order)
		if err != nil {
			log.Fatalln("Failed to marshal order:", err)
		}
		message := &sarama.ProducerMessage{
			Topic: KafkaTopic,
			Value: sarama.ByteEncoder(orderBytes),
		}
		partition, offset, err := producer.SendMessage(message)
		if err != nil {
			log.Fatalln("Failed to send message to Kafka:", err)
		}
		log.Printf("Message sent to partition %d at offset %d: %s\n", partition, offset, string(orderBytes))
	}
}

func generateOrders(lastPrice float64, symbol string) []*Order {
	rand.Seed(time.Now().UnixNano())

	// Half of the orders will be buy orders and half will be sell orders
	numOrders := 10 // You can change this to generate any number of orders
	halfNumOrders := numOrders / 2
	orders := make([]*Order, numOrders)

	// Generate buy orders with a price lower than lastPrice
	for i := 0; i < halfNumOrders; i++ {
		order := &Order{
			OrderID: i + 1,
			Side:    "Buy",
			Symbol:  symbol,
			Amount:  rand.Intn(1000),
			Price:   lastPrice - rand.Float64()*10,
		}
		orders[i] = order
	}

	// Generate sell orders with a price more than lastPrice
	for i := halfNumOrders; i < numOrders; i++ {
		order := &Order{
			OrderID: i + 1,
			Side:    "Sell",
			Symbol:  symbol,
			Amount:  rand.Intn(1000),
			Price:   lastPrice + rand.Float64()*10,
		}
		orders[i] = order
	}

	return orders
}
