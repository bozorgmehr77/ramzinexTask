package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func orderBookHandler(w http.ResponseWriter, r *http.Request) {

	symbol := r.URL.Query().Get("symbol")
	limitStr := r.URL.Query().Get("limit")
	var limit *float64
	if limitStr != "" {
		limitVal, err := strconv.ParseFloat(limitStr, 64)
		if err != nil {
			http.Error(w, "invalid limit parameter", http.StatusBadRequest)
			return
		}
		limit = &limitVal
	}

	// Call the GetOrderBook method to get the order book
	orderBook, err := getOrderBookFromDatabase(symbol, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Serialize the order book as JSON and write to the response
	jsonResp, err := json.Marshal(orderBook)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}

func runApi() {
	// Create a new HTTP server and register the handlers
	http.HandleFunc("/orders", orderBookHandler)
	log.Fatal(http.ListenAndServe(":8085", nil))
}
