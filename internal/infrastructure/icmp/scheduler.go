package icmp

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/websocket"
)

type ICMPMetricMessage struct {
	MessageType   string             `json:"messageType"`
	Timestamp     time.Duration      `json:"timestamp"`
	TimestampType string             `json:"timestampType"`
	Unit          string             `json:"unit"`
	Identifier    uint               `json:"identifier"`
	Data          *ICMPLatencyMetric `json:"data"`
}

type ICMPLatencyMetric struct {
	Metric  string `json:"metric"`
	Latency int64  `json:"latency"`
}

type Task struct {
	IpAddr string `json:"ip_addr"`
}

type TaskMetrics struct {
	Time   int64  `json:"time"`
	IpAddr string `json:"ip_addr"`
}

var protocolICMP = 1
var tasks = make(map[int]*Task)
var (
	// wg sync.WaitGroup
	mu sync.Mutex
)

func ICMPScheduler(ws *websocket.Conn, ch chan bool) error {
    for {
        mu.Lock()
        for key, task := range tasks {
            go func(key int, task *Task) {
                metric, err := icmpTask(task.IpAddr)
                if err != nil {
                    log.Printf("Error in ICMP Task ID %d: %v", key, err)
                } else {
                    jsonMetric, err := json.Marshal(&ICMPMetricMessage{
                        MessageType:   "icmp",
                        Timestamp:     time.Duration(time.Now().UnixMilli()),
                        TimestampType: "created_time",
                        Unit:          "host",
                        Identifier:    uint(key),
                        Data: &ICMPLatencyMetric{
                            Metric:  "ms",
                            Latency: metric.Time,
                        },
                    })
                    if err != nil {
                        log.Printf("Error marshaling JSON for task ID %d: %v", key, err)
                        return
                    }
                    _, err = ws.Write([]byte(jsonMetric))
                    if err != nil {
                        log.Printf("Error sending WebSocket message for task ID %d: %v", key, err)
                        ch <- true
                    }
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

	duration := time.Since(start).Milliseconds()
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
