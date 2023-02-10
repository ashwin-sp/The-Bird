package auth

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

func StartAuthServer() {
	fmt.Println("Auth GRPC up! ")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9292))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := Server{}
	grpcServer := grpc.NewServer()
	RegisterAuthServiceServer(grpcServer, &s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}
