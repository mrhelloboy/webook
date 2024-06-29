package ioc

import (
	"github.com/mrhelloboy/wehook/interactive/grpc"
	"github.com/mrhelloboy/wehook/pkg/grpcx"
	"github.com/mrhelloboy/wehook/pkg/logger"
	"github.com/spf13/viper"
	ggrpc "google.golang.org/grpc"
)

func InitGRPCxServer(l logger.Logger, intrServer *grpc.InteractiveServiceServer) *grpcx.Server {
	type Config struct {
		Port      int      `yaml:"port"`
		EtcdAddrs []string `yaml:"etcdAddrs"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	server := ggrpc.NewServer()
	intrServer.Register(server)

	return &grpcx.Server{
		Server:    server,
		Port:      cfg.Port,
		EtcdAddrs: cfg.EtcdAddrs,
		Name:      "interactive",
		L:         l,
	}
}
