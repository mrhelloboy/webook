package main

import (
	"log"
	"net"

	"github.com/mrhelloboy/wehook/api/proto/gen/intr/v1"
	"github.com/mrhelloboy/wehook/interactive/grpc"
	ggrpc "google.golang.org/grpc"
)

func main() {
	server := ggrpc.NewServer()
	// 注册你的服务
	intrSvc := &grpc.InteractiveServiceServer{}
	intrv1.RegisterInteractiveServiceServer(server, intrSvc)

	l, err := net.Listen("tcp", ":8090")
	if err != nil {
		panic(err)
	}

	err = server.Serve(l)
	log.Println(err)
}
