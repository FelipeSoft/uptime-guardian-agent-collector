package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/FelipeSoft/uptime-guardian-agent-collector/internal/infrastructure/network"
	pb "github.com/FelipeSoft/uptime-guardian-agent-collector/internal/uptime/v1/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	var targetSubnet string
	var targetProxyServer string
	var targetAgentId string

	fmt.Println("Describe which subnet the host belongs to (e.g. 192.168.120.1/24): ")
	fmt.Scan(&targetSubnet)

	fmt.Println("Describe the Proxy Server IPv4 Address and Port (e.g., 192.168.0.4:50051): ")
	fmt.Scan(&targetProxyServer)

	fmt.Println("Describe the Agent ID: ")
	fmt.Scan(&targetAgentId)

	os.Setenv("PROXY_SERVER", targetProxyServer)
	os.Setenv("SUBNET", targetSubnet)

	proxyAddr := os.Getenv("PROXY_SERVER")
	if proxyAddr == "" {
		log.Fatalf("Environment variable PROXY_SERVER is missing value")
	}

	subnet := os.Getenv("SUBNET")
	if subnet == "" {
		log.Fatalf("Environment variable SUBNET is missing value")
	}

	fmt.Printf("Informed PROXY_SERVER: %s \n", proxyAddr)
	fmt.Printf("Informed SUBNET: %s \n", subnet)
	fmt.Printf("Informed AGENT ID: %s \n", targetAgentId)

	// Validate gRPC server address
	log.Printf("Attempting to connect to gRPC server at %s", proxyAddr)

	// Use grpc.Dial (updated function)
	conn, err := grpc.NewClient(proxyAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Error connecting to gRPC server: %s", err.Error())
	}
	defer conn.Close()

	networkIPv4, err := network.GetLocalIPv4InSubnet(subnet)
	if err != nil {
		log.Printf("Error searching network interface adapters: %s", err.Error())
	}

	if len(networkIPv4) == 0 {
		log.Fatalf("No IPv4 addresses found in subnet: %s", subnet)
	}

	ipv4 := networkIPv4[0].String()

	client := pb.NewUptimeServiceClient(conn)
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)

			sentClientTime := time.Now().UnixMilli()

			res, err := client.SendCollectedData(ctx, &pb.UptimeRequest{
				HostId:   targetAgentId,
				SentTime: timestamppb.New(time.Now()),

				// should be put by array if have other IPs
				Ipv4: ipv4,
			})
			if err != nil {
				log.Printf("Error sending gRPC request: %s", err.Error())
			} else {
				receivedServerTime := res.SentTime.AsTime().UnixMilli()
				latency := receivedServerTime - sentClientTime
				fmt.Printf("Latency: %dms \n", latency)
			}
			cancel()
		}
	}()

	select {}
}
