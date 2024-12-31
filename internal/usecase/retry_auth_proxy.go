package usecase

import (
	"log"
	"os"
	"sync"
	"time"
)

type RetryProxyAuthUseCase struct {
	proxyAuthOutput   *ProxyAuthOutput
	wg                *sync.WaitGroup
	mu                *sync.Mutex
	attemptsLimit     int
	attemptDelay      time.Duration
	refreshTokenDelay time.Duration
	retryChannel      chan bool
}

func NewRetryProxyAuthUseCase(proxyAuthOutput *ProxyAuthOutput, wg *sync.WaitGroup, mu *sync.Mutex, attemptsLimit int, attemptDelay time.Duration, refreshTokenDelay time.Duration, retryChannel chan bool) *RetryProxyAuthUseCase {
	return &RetryProxyAuthUseCase{
		proxyAuthOutput:   proxyAuthOutput,
		wg:                wg,
		mu:                mu,
		attemptsLimit:     attemptsLimit,
		attemptDelay:      attemptDelay,
		refreshTokenDelay: refreshTokenDelay,
		retryChannel:      retryChannel,
	}
}

func (r *RetryProxyAuthUseCase) Execute() error {
	for {
		for attempt := 1; attempt < r.attemptsLimit; attempt++ {
			log.Printf("%d of %d Proxy Authentication with WebSocket Gateway\n", attempt, r.attemptsLimit)
			out, err := AuthProxy(ProxyAuthInput{
				Host:     os.Getenv("WEBSOCKET_GATEWAY_HTTP"),
				Protocol: "http",
				Path:     "/auth/proxy",
			})
			if err != nil {
				return err
			}
			if err == nil && out != nil && out.Token != "" {
				r.mu.Lock()
				*r.proxyAuthOutput = *out
				r.mu.Unlock()
				log.Println("Proxy authenticated successfully.")
				r.retryChannel <- false
				return nil
			}
			time.Sleep(r.attemptDelay)
		}

		log.Println("Reached max authentication attempts. Retrying after delay...")
		time.Sleep(r.refreshTokenDelay)
	}
}
