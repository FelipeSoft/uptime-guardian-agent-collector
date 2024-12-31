package main

import (
	"fmt"
	"github.com/FelipeSoft/uptime-guardian-agent-collector/internal/infrastructure/icmp"
	"github.com/FelipeSoft/uptime-guardian-agent-collector/internal/uptime"
	pb "github.com/FelipeSoft/uptime-guardian-agent-collector/internal/uptime/v1/proto"
	"github.com/FelipeSoft/uptime-guardian-agent-collector/internal/usecase"
	"github.com/joho/godotenv"
	"golang.org/x/net/websocket"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/url"
	"os"
	"sync"
	"time"
)

func main() {
	godotenv.Load("./../../.env")

	proxyUrl := (&url.URL{
		Scheme: "http",
		Host:   os.Getenv("PROXY_SERVER"),
	}).String()

	attemptsLimit := 5
	attemptDelay := 5 * time.Second
	refreshTokenDelay := 5 * time.Second

	var mu sync.Mutex
	var wg sync.WaitGroup

	retryChannel := make(chan bool)
	proxyAuthOutput := &usecase.ProxyAuthOutput{}
	retryAuthProxy := usecase.NewRetryProxyAuthUseCase(proxyAuthOutput, &wg, &mu, attemptsLimit, attemptDelay, refreshTokenDelay, retryChannel)

	go func() {
		for retry := range retryChannel {
			if retry {
				log.Println("retry for authentication...")
				retryAuthProxy.Execute()
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		retryAuthProxy.Execute()
	}()
	wg.Wait()

	mu.Lock()
	if proxyAuthOutput.Token == "" {
		mu.Unlock()
		log.Fatal("Failed to authenticate the proxy. Token is unavailable.")
	}
	token := proxyAuthOutput.Token
	mu.Unlock()

	wsUrl := &url.URL{
		Scheme:   "ws",
		Host:     os.Getenv("WEBSOCKET_GATEWAY_WS"),
		RawQuery: fmt.Sprintf("token=%s", token),
	}
	ws, err := websocket.Dial(wsUrl.String(), "", proxyUrl)
	if err != nil {
		log.Fatalf("WebSocket Gateway connection failed: %s", err.Error())
	}
	defer ws.Close()

	if err != nil {
		log.Fatalf("Failed on establishing connection with WebSocket Gateway.")
	}

	if err := icmp.AddICMPTask(1, &icmp.Task{IpAddr: "192.168.0.1"}); err != nil {
		log.Fatalf("Failed to add task 1: %v", err)
	}

	go func() {
		if err := icmp.ICMPScheduler(ws, retryChannel); err != nil {
			log.Fatalf("Scheduler error: %v", err)
		}
	}()

	go func() {
		for {
			var message string
			err := websocket.Message.Receive(ws, &message)
			if err != nil {
				log.Printf("Error on read message from WebSocket Gateway: %s", err)
				break
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

	log.Printf("\n gRPC server is available on %s", proxyUrl)
	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC server listening failed on %s caused by error: %s", proxyUrl, err.Error())
	}

	select {}
}
