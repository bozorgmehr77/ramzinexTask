## Order Book Subsystem

This project is a design and implementation of an order book subsystem.

### How to Prepare and Run the Project

1. Clone the repository.
2. Install the required dependencies:
   ````
   go mod vendor
   ```
3. Build the project:
   ````
   go build -ldflags="-s -w" -mod=vendor -o ramzinexTask.exe
   ```
4. Ensure that MySQL is installed and working correctly.
5. Run ZooKeeper and Kafka.
6. Create the Kafka topic "OrderBook":
   ````
   bin\windows\kafka-topics.bat --create --topic OrderBook --bootstrap-server localhost:9092
   ```
7. Run the application:
   ````
   ./ramzinexTask
   ```
8. To use the service, use an HTTP client to request the desired order book data, such as: `http://127.0.0.1:8085/orders?symbol=BTCUSDT&limit=5`.

### Code and File Structures

The project's code and file structures are as follows:

1. `common.go`: includes structs and helper functions.
2. `kafkaProducer.go`: creates random orders and writes them to Kafka. The value of the last price of each symbol is hard-coded. It creates 10 records for each symbol, with half being buy orders and the other half being sell orders. The range of orders is plus or minus 10.
3. `OrderBookDatabaseManager.go`: manages all connections with the database and all CRUD operations with the database.
4. `OrderBookManager.go`: manages all the behavior of the automated process of getting new orders from Kafka and processing them and putting them into the database in cooperation with `OrderBookDatabaseManager.go`.
5. `ApiManager.go`: creates an API service and runs an HTTP server to respond to clients about orders. It gets order data from the database in cooperation with `OrderBookDatabaseManager.go`.
6. `mainApp.go`: the main file that runs three goroutines: a Kafka producer (which runs every 20 seconds and creates 30 records for 3 symbols), an API (HTTP server), and `OrderBookManager`, which listens to Kafka, processes data, and writes them to the database.

### Database Structure and Tables

The database used is MySQL running on localhost with user root and no password. The application creates the database and table.

- Database name: task
- Table name: order
- Columns: order_id, side, symbol, amount, price
- order_id is the primary key, auto-increment, and unique.

### Structs

The following structs are used in the project:

```
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
```

### HTTP Services

The HTTP service listens on localhost:8085/orders. It receives an HTTP request with two GET parameters:

1. The `symbol` parameter is mandatory and can be "BTCUSDT", "BTCETH", or "BTCIRT".
2. The `limit` parameter is an optional integer that shows the maximum distance between the returned orders' price and the last price (which is hard-coded).

The service returns a JSON format that includes the last ID, asks, and bids in a sorted way.

### Performance Issue

The application does not use any performance-related algorithms. The only notable aspect of the code is how it handles the process of reading from Kafka and writing to the database. It starts a reverse timer with a value of 2 seconds. The code reads a record when it writes in the Kafka topic and puts it in a local list (cache), then it writes this list into the database when the length of this list becomes 10, or the timer becomes zero. The timer resets whenever a write in the database has occurred.

Every time a user requests an order book, the service retrieves the required data from the database. We do not have measurements of how time-consuming these queries are, but we can analyze the performance to see how well this code works in terms of scalability. However, this depends on factors beyond the code, such as the volume of data and the database used.
