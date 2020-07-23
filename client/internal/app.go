package app

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	pb "./proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// PlateList - a list of plates
type PlateList []pb.PlateRequest

// StartServerConn accepts the address, server conneciton timeout params and establish a connection to gRPC server
func StartServerConn(address string, serverConnTimeoutSecs int) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(serverConnTimeoutSecs)*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, address, grpc.WithBlock(), grpc.WithInsecure(), grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             100 * time.Millisecond,
		PermitWithoutStream: true,
	}))

	return conn, err
}

// SendPlates accepts the gRPC server connection, list of plates and deadline seconds.
// Plate data is sent as a stream to the gRPC server then receives a confirmation response.
func SendPlates(conn *grpc.ClientConn, plates PlateList, deadlineSecs int) error {
	total := len(plates)
	log.Printf("Client sending %v Plates\n", total)

	client := pb.NewPlatesClient(conn)

	deadlineMs := deadlineSecs * 1000
	clientDeadline := time.Now().Add(time.Duration(deadlineMs) * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)
	defer cancel()

	stream, err := client.SendPlates(ctx)
	if err != nil {
		return fmt.Errorf("Stream init SendPlates error: %v", err)
	}
	for _, pr := range plates {
		if err := stream.Send(&pr); err != nil {
			log.Printf("Stream Send error (%v) for plate: %+v \n", err, pr)
		}
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("Stream CloseAndRecv error %v, want %v", stream, err)
	}
	log.Printf("Server response: %v\n", reply.GetMessage())

	return nil
}

// ReadDataFile accepts a filename path then reads each line (comma delimited) and extracts each column and
// populates a request message per line then returns a list of plate data.
func ReadDataFile(filename string) (PlateList, error) {
	var pList PlateList

	// Open the file
	log.Printf("Reading file: %v\n", filename)
	file, err := os.Open(filename)
	if err != nil {
		return pList, err
	}

	// Setup file reader
	r := csv.NewReader(file)
	data, _ := r.ReadAll()

	// iterate each file line
	for i := range data {
		// extract columns
		column := data[i]

		// skip first line headers
		if i == 0 {
			continue
		}

		// populate columns
		pr := pb.PlateRequest{}

		if len(column[0]) > 0 {
			num, _ := strconv.Atoi(column[0])
			pr.Plate = int32(num)
		}
		if len(column[1]) > 1 {
			pr.Well = strings.TrimSpace(column[1])
		}
		if len(column[2]) > 2 {
			num, _ := strconv.Atoi(column[2])
			pr.Runtime = int32(num)
		}
		if len(column[3]) > 3 {
			fnum, _ := strconv.ParseFloat(column[3], 64)
			pr.OpticalDensity = float32(fnum)
		}
		if len(column[4]) > 4 {
			pr.Run = strings.TrimSpace(column[4])
		}
		if len(column[5]) > 5 {
			fnum, _ := strconv.ParseFloat(column[5], 64)
			pr.CorrectedOpticalDensity = float32(fnum)
		}

		pList = append(pList, pr)
	}

	return pList, file.Close()
}
