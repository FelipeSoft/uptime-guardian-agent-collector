package main

import (
	"context"
	"fmt"
	"github.com/FelipeSoft/uptime-guardian-agent-collector/internal/infrastructure/network"
	pb "github.com/FelipeSoft/uptime-guardian-agent-collector/internal/uptime/v1/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"os"
	"time"
)

func main() {
	var targetSubnet string
	var targetProxyServer string

	fmt.Println("Describe which subnet the host belongs to (e.g. 192.168.120.1/24): ")
	fmt.Scan(&targetSubnet)

	fmt.Println("Describe the Proxy Server IPv4 Address: ")
	fmt.Scan(&targetProxyServer)

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

	conn, err := grpc.NewClient(proxyAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Error on gRPC client declaring: %s", err.Error())
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	networkIPv4, err := network.GetLocalIPv4InSubnet(subnet)
	if err != nil {
		log.Printf("error on search network interface adapters: %s", err.Error())
	}

	client := pb.NewUptimeServiceClient(conn)
	sentTime := timestamppb.New(time.Now())
	client.SendCollectedData(ctx, &pb.UptimeRequest{
		ProxyServer:      proxyAddr,
		SentTime:         sentTime,
		Ipv4:             string(networkIPv4[0]),
		MacAddr:          "testing",
		OptionalHostname: &pb.UptimeRequest_Hostname{Hostname: "felip"},
		OptionalIpv6:     &pb.UptimeRequest_Ipv6{Ipv6: "felip"},
	})
}
