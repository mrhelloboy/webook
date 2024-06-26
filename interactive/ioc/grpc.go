package ioc

import (
	"github.com/mrhelloboy/wehook/interactive/grpc"
	"github.com/mrhelloboy/wehook/pkg/grpcx"
	"github.com/spf13/viper"
	ggrpc "google.golang.org/grpc"
)

func InitGRPCxServer(intrServer *grpc.InteractiveServiceServer) *grpcx.Server {
	type Config struct {
		Addr string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	server := ggrpc.NewServer()
	intrServer.Register(server)

	return &grpcx.Server{
		Server: server,
		Addr:   cfg.Addr,
	}
}
