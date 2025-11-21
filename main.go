// Dup2 prints the count and text of lines that appear more than once
// in the input. It reads from stdin or from a list of named files.
package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"google.golang.org/grpc"
)

var serverPorts = []int{4001, 4002, 4003, 4004, 4005}

func main() {
	if len(os.Args) < 2 {
		panic("Please provide a port number as a command-line argument.")
	}
	portIndex, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Printf("Error converting port: %v\n", err)
		return
	}
	if portIndex < 0 || portIndex >= len(serverPorts) {
		panic("Invalid port number.")
	}
	port := serverPorts[portIndex]
	fmt.Printf("Starting server on port: %d\n", port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Printf("Failed to listen: %v\n", err)
		return
	}
	grpcServer := grpc.NewServer()
	// Register your gRPC services here

	if err := grpcServer.Serve(lis); err != nil {
		fmt.Printf("Failed to serve: %v\n", err)
	}
}
