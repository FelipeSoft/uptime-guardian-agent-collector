package icmp

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type Task struct {
	IpAddr string `json:"ip_addr"`
}

type TaskMetrics struct {
	Time   time.Duration `json:"time"`
	IpAddr string        `json:"ip_addr"`
}

var protocolICMP = 1
var tasks = make(map[int]*Task)
var (
	// wg sync.WaitGroup
	mu sync.Mutex
)

func ICMPScheduler() error {
	for {
		mu.Lock()
		for key, task := range tasks {
			// fmt.Printf("ICMP Task ID: %d; IP Address: %s \n", key, task.IpAddr)
			go func(key int, task *Task) {
				metric, err := icmpTask(task.IpAddr)
				if err != nil {
					// emit Error Event
					// messages:
					// message type: destination unreachable; read ip4 0.0.0.0 i/o timeout
					fmt.Printf("Error in ICMP Task ID %d for IP %s: %v\n", key, task.IpAddr, err)
				} else {
					// emit Success Event
					fmt.Printf("ICMP Task ID: %d IP %s; Time: %s completed successfully.\n", key, metric.IpAddr, metric.Time)
				}
			}(key, task)
		}
		mu.Unlock()
		time.Sleep(1 * time.Second)
	}
}

func AddICMPTask(key int, task *Task) error {
	mu.Lock()
	defer mu.Unlock()

	if _, exists := tasks[key]; exists {
		return fmt.Errorf("cannot use the unique key %d because it already exists", key)
	}

	tasks[key] = task
	return nil
}

func UpdateICMPTask(key int, task *Task) error {
	mu.Lock()
	defer mu.Unlock()

	if _, exists := tasks[key]; !exists {
		return fmt.Errorf("the provided task id %d does not exists", key)
	}

	tasks[key] = task
	return nil
}

func RemoveICMPTask(unique_key int) {
	mu.Lock()
	defer mu.Unlock()
	delete(tasks, unique_key)
}

func icmpTask(addr string) (*TaskMetrics, error) {
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, err
	}
	defer c.Close()

	msg := &icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: []byte(""),
		},
	}
	wb, err := msg.Marshal(nil)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	if _, err := c.WriteTo(wb, &net.IPAddr{IP: net.ParseIP(addr)}); err != nil {
		return nil, err
	}

	reply := make([]byte, 1500)
	if err = c.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return nil, err
	}

	n, peer, err := c.ReadFrom(reply)
	if err != nil {
		return nil, err
	}

	duration := time.Since(start)
	msg, err = icmp.ParseMessage(protocolICMP, reply[:n])
	if err != nil {
		return nil, fmt.Errorf("failed to parse ICMP reply: %w", err)
	}

	switch msg.Type {
	case ipv4.ICMPTypeEchoReply:
		return &TaskMetrics{Time: duration, IpAddr: peer.String()}, nil
	default:
		return nil, fmt.Errorf("unexpected ICMP message type: %v", msg.Type)
	}
}
