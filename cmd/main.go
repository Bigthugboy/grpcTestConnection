package main

import (
	"google.golang.org/grpc"
	"grpcTestConnection/server/payment/server/payment"
	"grpcTestConnection/server/server"
	"log"
	"net"
)

func main() {
	lis, err := net.Listen("tcp", ":50081")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	paymentServer := server.NewServer()

	payment.RegisterAddDebitAccountServiceServer(grpcServer, paymentServer)

	log.Println("gRPC server is running on port 50081")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
