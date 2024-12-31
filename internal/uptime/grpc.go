package uptime

import (
	"context"
	"encoding/json"
	"strconv"
	"time"
	pb "github.com/FelipeSoft/uptime-guardian-agent-collector/internal/uptime/v1/proto"
	"golang.org/x/net/websocket"
)

type UptimeService struct {
	pb.UnimplementedUptimeServiceServer
	ws *websocket.Conn
	ch chan bool
}

type AgentMetricMessage struct {
	MessageType   string              `json:"messageType"`
	Timestamp     time.Duration       `json:"timestamp"`
	TimestampType string              `json:"timestampType"`
	HostId        int                 `json:"hostId"`
	Addr          string              `json:"addr"`
	Data          *AgentLatencyMetric `json:"data"`
}

type AgentLatencyMetric struct {
	Metric  string `json:"metric"`
	Latency int64  `json:"latency"`
}

func NewUptimeService(ws *websocket.Conn, ch chan bool) *UptimeService {
	return &UptimeService{
		ws: ws,
		ch: ch,
	}
}

func (s *UptimeService) SendCollectedData(ctx context.Context, message *pb.UptimeRequest) (*pb.UptimeResponse, error) {
	sentTime := message.SentTime.AsTime().UnixMilli()
	receivedTime := time.Now().UnixMilli()

	intHostId, err := strconv.Atoi(message.HostId)
	if err != nil {
		return nil, err
	}

	metric := &AgentMetricMessage{
		MessageType:   "agent",
		Timestamp:     time.Duration(time.Now().UnixMilli()),
		TimestampType: "created_time",
		HostId:        intHostId,
		Addr:          message.Ipv4,
		Data: &AgentLatencyMetric{
			Metric:  "ms",
			Latency: receivedTime - sentTime,
		},
	}

	msg, err := json.Marshal(metric)
	if err != nil {
		return nil, err
	}

	_, err = s.ws.Write([]byte(msg))
	if err != nil {
		return nil, err
	}

	return &pb.UptimeResponse{
		SentTime: message.SentTime,
		HostId:   message.HostId,
		Ipv4:     message.Ipv4,
	}, nil
}
