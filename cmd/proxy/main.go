package main

import (
	"fmt"
	"log"
	"github.com/FelipeSoft/uptime-guardian-agent-collector/internal/infrastructure/icmp"
)

// "log"
// "net"

// "github.com/FelipeSoft/uptime-guardian-agent-collector/internal/uptime"
// pb "github.com/FelipeSoft/uptime-guardian-agent-collector/internal/uptime/v1/proto"
// "google.golang.org/grpc"

func main() {
	if err := icmp.AddICMPTask(1, &icmp.Task{IpAddr: "192.168.200.153"}); err != nil {
		log.Fatalf("Failed to add task 1: %v", err)
	}
	if err := icmp.AddICMPTask(2, &icmp.Task{IpAddr: "192.168.200.152"}); err != nil {
		log.Fatalf("Failed to add task 2: %v", err)
	}
	// if err := icmp.AddICMPTask(3, &icmp.Task{IpAddr: "189.124.132.32"}); err != nil {
	// 	log.Fatalf("Failed to add task 3: %v", err)
	// }
	// if err := icmp.AddICMPTask(4, &icmp.Task{IpAddr: "189.124.132.33"}); err != nil {
	// 	log.Fatalf("Failed to add task 4: %v", err)
	// }

	go func() {
		if err := icmp.ICMPScheduler(); err != nil {
			log.Fatalf("Scheduler error: %v", err)
		}
	}()

	fmt.Println("Updated Motorola Edge 30 Task IP Address!")

	select {}
	// (Monitoramento Ativo) Receber do agente as informações
	// (Monitoramento Passivo) -> se o hardware do Proxy for um pouco mais robusto ICMP com os hosts
	// lis, err := net.Listen("tcp", "192.168.200.154:50051")
	// if err != nil {
	// 	log.Fatalf("listening to gRPC failed with: %s", err.Error())
	// }
	// defer lis.Close()
	// grpcServer := grpc.NewServer()
	// pb.RegisterUptimeServiceServer(grpcServer, &uptime.UptimeService{})
	// log.Println("gRPC server is running on port 50051")
	// if err = grpcServer.Serve(lis); err != nil {
	// log.Fatalf("listening on TCP 50051 gRPC port failed: %s", err.Error())
	// }
}
