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
		OptionalIpv6:     &pb.UptimeResponse_Ipv6{Ipv6: message.GetIpv6()},
		MacAddr:          message.MacAddr,
		OptionalHostname: &pb.UptimeResponse_Hostname{Hostname: message.GetHostname()},
	}, nil
}
