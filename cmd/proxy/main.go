package main

import (
	"fmt"
	"log"
	"net"

	"github.com/FelipeSoft/uptime-guardian-agent-collector/internal/infrastructure/icmp"
	"github.com/FelipeSoft/uptime-guardian-agent-collector/internal/uptime"
	pb "github.com/FelipeSoft/uptime-guardian-agent-collector/internal/uptime/v1/proto"
	"golang.org/x/net/websocket"
	"google.golang.org/grpc"
)

func main() {
	url := "http://localhost:9000"
	origin := "http://localhost:50051"
	ws, err := websocket.Dial(url, "", origin)
	defer ws.Close()

	if err != nil {
		log.Fatalf("Failed on establishing connection with Websocket Gateway.")
	}

	if err := icmp.AddICMPTask(1, &icmp.Task{IpAddr: "192.168.200.153"}); err != nil {
		log.Fatalf("Failed to add task 1: %v", err)
	}

	go func() {
		if err := icmp.ICMPScheduler(); err != nil {
			log.Fatalf("Scheduler error: %v", err)
		}
	}()

	go func () {
		for {
			var message string
			err := websocket.Message.Receive(ws, &message)
			if err != nil {
				log.Printf("Error on read message from Websocket Gateway: %s", err)
			}
			fmt.Println(message)
		}
	}()

	// (Monitoramento Ativo) Receber do agente as informações
	// (Monitoramento Passivo) -> se o hardware do Proxy for um pouco mais robusto ICMP com os hosts
	lis, err := net.Listen("tcp", "192.168.200.154:50051")
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
