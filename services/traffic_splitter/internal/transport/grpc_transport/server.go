package grpctransport

import (
	"fmt"
	"net"

	splitterv1 "github.com/tahion/ml_ab_test_service/gen/go/splitter/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	grpcServer *grpc.Server
	port       int
}

func NewServer(port int, service splitterv1.TrafficSplitterServer) *Server {
	s := grpc.NewServer()
	splitterv1.RegisterTrafficSplitterServer(s, service)
	reflection.Register(s)
	return &Server{grpcServer: s, port: port}

}

func (s *Server) Run() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	return s.grpcServer.Serve(lis)
}

func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
}
