package main

import (
	"log"
	"net"
	"github.com/FelipeSoft/uptime-guardian-agent-collector/internal/uptime"
	pb "github.com/FelipeSoft/uptime-guardian-agent-collector/internal/uptime/v1/proto"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", "127.0.0.1:50051")
	if err != nil {
		log.Fatalf("listening to gRPC failed with: %s", err.Error())
	}	
	defer lis.Close()
	grpcServer := grpc.NewServer()
	pb.RegisterUptimeServiceServer(grpcServer, &uptime.UptimeService{})
	log.Println("gRPC server is running on port 50051")
	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalf("listening on TCP 50051 gRPC port failed: %s", err.Error())
	}
}