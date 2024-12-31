package uptime

import (
	"context"
	// "fmt"

	pb "github.com/FelipeSoft/uptime-guardian-agent-collector/internal/uptime/v1/proto"
)

type UptimeService struct {
	pb.UnimplementedUptimeServiceServer
}

func (s *UptimeService) SendCollectedData(ctx context.Context, message *pb.UptimeRequest) (*pb.UptimeResponse, error) {
	// fmt.Println(message)
	
	return &pb.UptimeResponse{
		SentTime: message.SentTime,
		HostId:   message.HostId,
		Ipv4:     message.Ipv4,
	}, nil
}
