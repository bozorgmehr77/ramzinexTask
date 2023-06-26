package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"sort"
)

// connectToDatabase connects to the MySQL server and returns a database connection.
func connectToDatabase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// createTaskDatabase creates the `task` database if it does not already exist.
func createTaskDatabase(db *sql.DB) error {
	_, err := db.Exec("CREATE DATABASE IF NOT EXISTS `task`")
	if err != nil {
		return err
	}
	return nil
}

// createOrderTable creates the `order` table if it does not already exist.
func createOrderTable(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS `task`.`order` (" +
		"`order_id` INT(11) NOT NULL AUTO_INCREMENT," + // Add AUTO_INCREMENT here
		"`side` ENUM('buy', 'sell') NOT NULL," +
		"`symbol` VARCHAR(10) NOT NULL," +
		"`amount` INT(11) NOT NULL," +
		"`price` DOUBLE NOT NULL," +
		"PRIMARY KEY (`order_id`)" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;")
	if err != nil {
		return err
	}
	return nil
}

// addOrderBookRecords adds new order objects to the `order` table.
// It checks if an order with the same order_id already exists in the table,
// and only inserts new orders that do not already exist.
func addOrderBookRecords(db *sql.DB, orders []Order) error {
	for _, order := range orders {
		// Insert the new order into the `order` table.
		result, err := db.Exec("INSERT INTO `task`.`order` (`side`, `symbol`, `amount`, `price`) VALUES (?, ?, ?, ?)",
			order.Side, order.Symbol, order.Amount, order.Price)
		if err != nil {
			return err
		}

		orderID, err := result.LastInsertId()
		if err != nil {
			return err
		}
		order.OrderID = int(orderID)

		fmt.Printf("Added order: %+v\n", order)
	}
	return nil
}

func getOrdersBySide(db *sql.DB, limit *float64, symbol string, side string) ([]Order, int, error) {
	// Get the last price of the symbol from the ?
	lastPrice, err := getLastPrice(symbol)
	if err != nil {
		return nil, 0, err
	}
	// Build the SQL query to get the orders
	query := "SELECT `order_id`, `side`, `symbol`, `amount`, `price` FROM `task`.`order` WHERE `symbol` = ? AND `side` = ?"
	var args []interface{}
	args = append(args, symbol, side)
	if limit != nil && *limit > 0 {
		query += " AND (price <= ? AND price >= ?)"
		args = append(args, lastPrice+*limit, lastPrice-*limit)
	}
	// Execute the query and process the results.
	var orders []Order
	var lastUpdateID int
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var order Order
		err := rows.Scan(&order.OrderID, &order.Side, &order.Symbol, &order.Amount, &order.Price)
		if err != nil {
			return nil, 0, err
		}
		orders = append(orders, order)
		lastUpdateID = order.OrderID
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// Return the generated orders and the last update ID.
	return orders, lastUpdateID, nil
}

func getOrderBookFromDatabase(symbol string, limit *float64) (OrderBook, error) {
	//Connect to database:
	db, err := connectToDatabase("root:@tcp(localhost:3306)/")
	if err != nil {
		log.Fatal(err)
	}

	// Get the buy and sell orders from the database
	buyOrders, lastUpdateID, err := getOrdersBySide(db, limit, symbol, "BUY")
	sellOrders, _, err := getOrdersBySide(db, limit, symbol, "SELL")

	// Sort the buy and sell orders by price
	sort.Slice(buyOrders, func(i, j int) bool {
		return buyOrders[i].Price > buyOrders[j].Price
	})
	sort.Slice(sellOrders, func(i, j int) bool {
		return sellOrders[i].Price < sellOrders[j].Price
	})

	// Build the bids and asks arrays
	var bids [][]float64
	for _, order := range buyOrders {
		bids = append(bids, []float64{order.Price, float64(order.Amount)})
	}
	var asks [][]float64
	for _, order := range sellOrders {
		asks = append(asks, []float64{order.Price, float64(order.Amount)})
	}

	// Create the OrderBook struct and return it
	orderBook := OrderBook{
		LastUpdateID: lastUpdateID,
		Bids:         bids,
		Asks:         asks,
	}
	return orderBook, nil
}
