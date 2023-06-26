package main

func main() {

	go produceSamplesInLoop()
	go runApi()
	db, _ := initializeOrderBook()
	listenReadWrite(db)
}
