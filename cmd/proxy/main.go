package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/FelipeSoft/uptime-guardian-agent-collector/internal/infrastructure/icmp"
	"github.com/FelipeSoft/uptime-guardian-agent-collector/internal/uptime"
	pb "github.com/FelipeSoft/uptime-guardian-agent-collector/internal/uptime/v1/proto"
	"github.com/joho/godotenv"
	"golang.org/x/net/websocket"
	"google.golang.org/grpc"
)

func main() {
	godotenv.Load("./../../.env")
	websocketGatewayUrl := os.Getenv("WEBSOCKET_GATEWAY") + "?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwcm94eUlkIjoxLCJpYXQiOjE3MzU0ODkxNDIsImV4cCI6MTczNTQ4OTc0Mn0.DPnhBss05IK3hU4Ldg6Gl8C-ZcNjBgswaoaufZkFJhA"
	origin := "http://192.168.0.4:50051"

	ws, err := websocket.Dial(websocketGatewayUrl, "", origin)
	if err != nil {
		log.Fatalf("WebSocket Gateway connection failed: %s", err.Error())
	}
	defer ws.Close()

	if err != nil {
		log.Fatalf("Failed on establishing connection with WebSocket Gateway.")
	}

	if err := icmp.AddICMPTask(1, &icmp.Task{IpAddr: "192.168.0.6"}); err != nil {
		log.Fatalf("Failed to add task 1: %v", err)
	}

	go func() {
		if err := icmp.ICMPScheduler(); err != nil {
			log.Fatalf("Scheduler error: %v", err)
		}
	}()

	go func() {
		for {
			var message string
			err := websocket.Message.Receive(ws, &message)
			if err != nil {
				log.Printf("Error on read message from WebSocket Gateway: %s", err)
			}
			fmt.Println(message)
		}
	}()

	// (Active Monitoring) Proxy receive agent informations
	// (Passive Monitoring) -> If the proxy hardware is robust, could make ICMP calls with hosts (installed agents)
	lis, err := net.Listen("tcp", os.Getenv("PROXY_SERVER"))
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

	select {}
}
