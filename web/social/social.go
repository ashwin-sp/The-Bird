package social

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

func StartSocialServer() {
	fmt.Println("Social GRPC up! ")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9595))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := Server{}
	grpcServer := grpc.NewServer()
	RegisterSocialServiceServer(grpcServer, &s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}
