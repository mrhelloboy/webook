package interceptors

import (
	"context"
	"net"
	"strings"

	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc/peer"
)

type Builder struct{}

// PeerName 获取对端应用名称
func (b *Builder) PeerName(ctx context.Context) string {
	return b.grpcHeaderValue(ctx, "app")
}

// PeerIP 获取对端 ip
func (b *Builder) PeerIP(ctx context.Context) string {
	// 如果在 ctx 里面传入。或者说客户端里面设置了，就直接用它设置的
	// 有些时候你经过网关之类，就需要客户端主动设置，防止后面拿到网关的 IP
	clientIP := b.grpcHeaderValue(ctx, "client-ip")
	if clientIP != "" {
		return clientIP
	}
	// 从grpc里取对端 ip
	pr, ok2 := peer.FromContext(ctx)
	if !ok2 {
		return ""
	}

	if pr.Addr == net.Addr(nil) {
		return ""
	}

	addSlice := strings.Split(pr.Addr.String(), ":")
	if len(addSlice) > 1 {
		return addSlice[0]
	}
	return ""
}

func (b *Builder) grpcHeaderValue(ctx context.Context, key string) string {
	if key == "" {
		return ""
	}
	// 如果要在 gprc 客户端和服务端之间传递元数据，就用这个
	// 服务端接收数据，读取数据的用法
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	return strings.Join(md.Get(key), ";")
}
