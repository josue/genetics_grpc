package main

import (
	"flag"
	"log"
	"os"
	"time"

	app "./internal"
)

func main() {
	start := time.Now()

	// capture CLI arguments
	var address = flag.String("server", "localhost:50051", "grpc server address")
	var serverConnTimeoutSecs = flag.Int("timeout", 5, "grpc server connection timeout in secs")
	var transTimeoutSecs = flag.Int("deadline", 60, "grpc trans timeout in secs")
	var filename = flag.String("file", "test/od-data.csv", "plates file data, comma-delimited")
	flag.Parse()

	// get server params
	serverAddr := *address
	if os.Getenv("GRPC_SERVER") != "" {
		serverAddr = os.Getenv("GRPC_SERVER")
	}

	// get filename to parse
	dataFile := *filename
	if os.Getenv("DATAFILE") != "" {
		dataFile = os.Getenv("DATAFILE")
	}

	// initiate gRPC server connection
	conn, err := app.StartServerConn(serverAddr, *serverConnTimeoutSecs)
	if err != nil {
		log.Printf("Server connection error: %v", err)
		log.Fatalf("Duration: %v", time.Since(start))
	}

	// read plates data from designated data file
	plates, err := app.ReadDataFile(dataFile)
	if err != nil {
		log.Printf("File error: %v", err)
		log.Fatalf("Duration: %v", time.Since(start))
	}

	// send plates data to gRPC connection
	err = app.SendPlates(conn, plates, *transTimeoutSecs)
	if err != nil {
		log.Printf("SendPlates error: %v", err)
		log.Fatalf("Duration: %v", time.Since(start))
	}

	log.Printf("Duration: %v", time.Since(start))

	if err = conn.Close(); err != nil {
		log.Printf("Server close connection error: %v", err)
		log.Fatalf("Duration: %v", time.Since(start))
	}
}
