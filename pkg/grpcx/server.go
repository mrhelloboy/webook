package grpcx

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/mrhelloboy/wehook/pkg/netx"

	"github.com/mrhelloboy/wehook/pkg/logger"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"

	"google.golang.org/grpc"
)

type Server struct {
	*grpc.Server
	Port      int
	EtcdAddrs []string
	Name      string
	L         logger.Logger
	KaCancel  func() // lease keep-alive cancel func
	em        endpoints.Manager
	client    *etcdv3.Client
	key       string
}

func (s *Server) Serve() error {
	l, err := net.Listen("tcp", ":"+strconv.Itoa(s.Port))
	if err != nil {
		return err
	}
	err = s.register()
	if err != nil {
		return err
	}
	return s.Server.Serve(l)
}

// 注册服务
func (s *Server) register() error {
	client, err := etcdv3.New(etcdv3.Config{
		Endpoints: s.EtcdAddrs,
	})
	if err != nil {
		return err
	}
	s.client = client

	em, err := endpoints.NewManager(client, "service/"+s.Name)
	s.em = em
	if err != nil {
		return err
	}
	addr := netx.GetOutboundIP() + ":" + strconv.Itoa(s.Port)
	key := "service/" + s.Name + "/" + addr
	s.key = key
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	var ttl int64 = 30
	leaseResp, err := client.Grant(ctx, ttl)

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 向注册中心注册服务实例定位信息 + 及添加租约
	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		Addr: addr,
	}, etcdv3.WithLease(leaseResp.ID))

	// 续租
	kaCtx, kaCancel := context.WithTimeout(context.Background(), time.Second)
	s.KaCancel = kaCancel
	ch, err := client.KeepAlive(kaCtx, leaseResp.ID)
	if err != nil {
		return err
	}
	go func() {
		for kaResp := range ch {
			// 打印一下 DEBUG 日志
			s.L.Debug(kaResp.String())
		}
	}()

	return nil
}

func (s *Server) Close() error {
	// 取消续租
	if s.KaCancel != nil {
		s.KaCancel()
	}

	// 删除服务实例
	if s.em != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := s.em.DeleteEndpoint(ctx, s.key)
		if err != nil {
			return err
		}
	}

	// 关闭客户端
	if s.client != nil {
		err := s.client.Close()
		if err != nil {
			return err
		}
	}

	// 关闭 gRPC 服务
	s.Server.GracefulStop()

	return nil
}
