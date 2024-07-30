package main

import (
	"google.golang.org/grpc"
	"grpcTestConnection/server/grpcserver"
	"grpcTestConnection/server/payment/grpcserver/payment"
	"log"
	"net"
)

func main() {
	const port = ":50081"

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	paymentServer := grpcserver.NewServer()

	payment.RegisterInternalsServiceServer(grpcServer, paymentServer)

	log.Println("gRPC grpcServer is running on port 50081")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
