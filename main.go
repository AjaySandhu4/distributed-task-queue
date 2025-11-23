// Dup2 prints the count and text of lines that appear more than once
// in the input. It reads from stdin or from a list of named files.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	pb "github.com/AjaySandhu4/distributed-task-queue/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type server struct {
	pb.UnimplementedGreeterServer
	serverIndex int
}

var serverPorts = []int{4001, 4002, 4003}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	fmt.Printf("Server %d Received: %v\n", s.serverIndex, in.GetName())
	return &pb.HelloResponse{Message: "Hello " + in.GetName() + " from server " + fmt.Sprintf("%d", s.serverIndex)}, nil
}

func startServer(serverIndex int, ctx context.Context) {
	port := serverPorts[serverIndex]
	fmt.Printf("Starting server on port: %d\n", port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Printf("Failed to listen: %v\n", err)
		return
	}
	grpcServer := grpc.NewServer()

	// Register your gRPC services here
	pb.RegisterGreeterServer(grpcServer, &server{serverIndex: serverIndex})

	// Start the server in a goroutine
	go func() {
		fmt.Printf("gRPC server listening on port %d\n", port)
		if err := grpcServer.Serve(lis); err != nil {
			fmt.Printf("Failed to serve: %v\n", err)
		}
	}()

	// Wait for context cancellation to gracefully stop the server
	<-ctx.Done()
	fmt.Printf("Shutting down server on port %d\n", port)
	grpcServer.GracefulStop()
}

func sayHelloToServer(targetIndex int, serverIndex int) {
	log.Println("Saying hello to server", targetIndex, "from server", serverIndex)
	addr := fmt.Sprintf("localhost:%d", serverPorts[targetIndex])
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("Failed to connect to server %d: %v\n", targetIndex, err)
		return
	}
	c := pb.NewGreeterClient(conn)
	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: strconv.Itoa(serverIndex)})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())
	conn.Close()
	cancel()
}

func main() {
	if len(os.Args) < 2 {
		panic("Please provide a server index as a command-line argument.")
	}
	serverIndex, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Printf("Error converting server index: %v\n", err)
		return
	}
	if serverIndex < 0 || serverIndex >= len(serverPorts) {
		panic("Invalid server index.")
	}

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start the server
	go startServer(serverIndex, ctx)

	time.Sleep(2 * time.Second) // Wait for all servers to start

	// Say hello to other servers in parallel
	for i := range serverPorts {
		if i != serverIndex {
			go sayHelloToServer(i, serverIndex)
		}
	}

	<-sigChan
	cancel()
	log.Print("End of main function reached for server ", serverIndex)
}
