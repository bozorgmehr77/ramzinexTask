package main

import (
	"database/sql"
	"fmt"
	"github.com/Shopify/sarama"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func initializeOrderBook() (*sql.DB, error) {
	db, err := connectToDatabase("root:@tcp(localhost:3306)/")
	if err != nil {
		log.Fatal(err)
	}
	//defer db.Close()

	err = createTaskDatabase(db)
	if err != nil {
		log.Fatal(err)
	}

	err = createOrderTable(db)
	if err != nil {
		log.Fatal(err)
	}

	return db, nil
}

func listenReadWrite(db *sql.DB) {
	// Create a new Kafka consumer.
	consumer, err := createKafkaConsumer()
	if err != nil {
		log.Fatal("Failed to create Kafka consumer:", err)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Println("Failed to close Kafka consumer cleanly:", err)
		}
	}()

	// Start consuming messages from Kafka.
	partitions, err := consumer.Partitions(KafkaTopic)
	if err != nil {
		log.Fatal("Failed to get partition list:", err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	orders := []Order{}
	writeTimer := time.NewTimer(2 * time.Second)
	defer writeTimer.Stop()
	for {
		select {
		case <-writeTimer.C:
			if len(orders) > 0 {
				err := addOrderBookRecords(db, orders)
				if err != nil {
					log.Fatal("Failed to add orders to database:", err)
				}
				log.Println("Orders added to database:", orders)
				orders = nil
			}
			writeTimer.Reset(2 * time.Second)
		case <-signals:
			fmt.Println("Exiting...")
			time.Sleep(time.Second)
			return
		default:
			for _, partition := range partitions {
				partitionConsumer, err := consumer.ConsumePartition(KafkaTopic, partition, sarama.OffsetNewest)
				if err != nil {
					log.Fatalln("Failed to start partition consumer:", err)
				}

				for {
					select {
					case msg := <-partitionConsumer.Messages():
						order, err := DecodeOrder(msg.Value)
						if err != nil {
							log.Println("Failed to decode order:", err)
							continue
						}
						orders = append(orders, order)
						log.Println("Order added to list:", order)

						if len(orders) == 10 {
							err := addOrderBookRecords(db, orders)
							if err != nil {
								log.Fatal("Failed to add orders to database:", err)
							}
							log.Println("Orders added to database:", orders)
							orders = nil
							writeTimer.Reset(2 * time.Second)
						}
					case err := <-partitionConsumer.Errors():
						log.Println("Partition consumer error:", err)
					default:
						time.Sleep(100 * time.Millisecond)
					}
				}
			}
		}
	}
}

func createKafkaConsumer() (sarama.Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer([]string{KafkaBroker}, config)
	if err != nil {
		return nil, err
	}

	return consumer, nil
}
