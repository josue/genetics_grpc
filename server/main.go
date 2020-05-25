package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	app "./internal"
	pb "./internal/proto"

	"google.golang.org/grpc"
)

func main() {
	// capture CLI arguments
	var port = flag.Int("port", 50051, "grpc server port")
	var output = flag.String("o", "stdout", "output redirection")
	flag.Parse()

	// get server params
	outputType := *output
	if os.Getenv("OUTPUT") != "" {
		outputType = os.Getenv("OUTPUT")
	}
	log.Printf("Output type: %v", outputType)

	// initialize TCP connection with designated port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", *port))
	if err != nil {
		log.Fatalf("Net listener failed: %v", err)
	}

	// App setup + init
	app := app.Server{OutputType: outputType}
	app.GetEnvConfig()
	app.Init()
	err = app.SetupDB()
	if err != nil {
		log.Fatalf("DB Setup failed: %v", err)
	}

	// gRPC serve init
	s := grpc.NewServer()
	pb.RegisterPlatesServer(s, &app)

	log.Printf("gRPC server starting on %v ...", *port)

	// serve gRPC connection
	if err := s.Serve(lis); err != nil {
		log.Fatalf("gRPC failed to serve: %v", err)
	}
}
