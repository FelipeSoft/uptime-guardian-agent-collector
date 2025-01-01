package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/FelipeSoft/uptime-guardian-agent-collector/internal/infrastructure/icmp"
	"github.com/FelipeSoft/uptime-guardian-agent-collector/internal/uptime"
	pb "github.com/FelipeSoft/uptime-guardian-agent-collector/internal/uptime/v1/proto"
	"github.com/FelipeSoft/uptime-guardian-agent-collector/internal/usecase"
	"github.com/joho/godotenv"
	"golang.org/x/net/websocket"
	"google.golang.org/grpc"
)

func main() {
	godotenv.Load("./../../.env")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
	grpcMessagesChannel := make(chan []byte, 10)

	proxyAuthOutput := &usecase.ProxyAuthOutput{}
	retryAuthProxy := usecase.NewRetryProxyAuthUseCase(proxyAuthOutput, &wg, &mu, attemptsLimit, attemptDelay, refreshTokenDelay, retryChannel)

	go func() {
		for {
			select {
			case retry, ok := <-retryChannel:
				if !ok {
					retryAuthProxy.Execute()
					return
				}
				if retry {
					log.Println("retry for authentication...")
					retryAuthProxy.Execute()
				}
			}
		}
	}()
	
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := retryAuthProxy.Execute(); err != nil {
			log.Printf("Error during initial proxy authentication: %v", err)
			retryChannel <- true
		}
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

	// if err := icmp.AddICMPTask(2, &icmp.Task{IpAddr: "192.168.0.7"}); err != nil {
	// 	log.Fatalf("Failed to add task 1: %v", err)
	// }

	go func() {
		if err := icmp.ICMPScheduler(ws, retryChannel); err != nil {
			log.Printf("ICMP scheduler error: %v", err)
			retryChannel <- true
		}
	}()

	go func() {
		for {
			var message string
			err := websocket.Message.Receive(ws, &message)
			if err != nil {
				log.Fatalf("WebSocket error: %v. Attempting reconnect...", err)
				retryChannel <- true
				break
			}
			fmt.Println(message)
		}
	}()

	go func() {
		for msg := range grpcMessagesChannel {
			_, err := ws.Write([]byte(msg))
			if err != nil {
				log.Printf("Error on send collected metrics from agent: %v", err)
				retryChannel <- true
			}
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

	uptimeService := uptime.NewUptimeService(ws, grpcMessagesChannel)
	pb.RegisterUptimeServiceServer(grpcServer, uptimeService)

	log.Printf("gRPC server is available on %s", proxyUrl)

	go func() {
		log.Println("Starting gRPC server...")
		if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			log.Fatalf("gRPC server terminated unexpectedly: %v", err)
		}
		log.Println("gRPC server has stopped.")
	}()

	go func() {
		<-ctx.Done()
		log.Println("Shutting down gracefully...")
		grpcServer.GracefulStop()
		log.Println("gRPC server stopped.")
		os.Exit(0)
	}()

	select {}
}
