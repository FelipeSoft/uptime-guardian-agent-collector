package uptime

import (
	"context"
	pb "github.com/FelipeSoft/uptime-guardian-agent-collector/internal/uptime/v1/proto"
)

type UptimeService struct {
	pb.UnimplementedUptimeServiceServer
}

func (s *UptimeService) SendCollectedData(ctx context.Context, message *pb.UptimeRequest) (*pb.UptimeResponse, error) {
	return &pb.UptimeResponse{
		SentTime:         message.SentTime,
		ProxyServer:      message.ProxyServer,
		Ipv4:             message.Ipv4,
	}, nil
}
