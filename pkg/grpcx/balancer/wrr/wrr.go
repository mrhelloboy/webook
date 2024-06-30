package wrr

import (
	"sync"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

// 自定义负载均衡算法需要实现接口：
// base.PickerBuilder 接口
// balancer.Picker 接口

const name = "custom_wrr"

func init() {
	balancer.Register(base.NewBalancerBuilder(name, &wrrPickerBuilder{}, base.Config{HealthCheck: false}))
}

type wrrPickerBuilder struct{}

func (*wrrPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*conn, 0, len(info.ReadySCs))
	// sc: subConn
	// sci: subConnInfo
	for sc, sci := range info.ReadySCs {
		cc := &conn{
			cc: sc,
		}
		// 节点的权重
		md, ok := sci.Address.Metadata.(map[string]any)
		if ok {
			weightVal := md["weight"]
			if weight, ok := weightVal.(float64); ok {
				cc.weight = int(weight)
			}
		}

		if cc.weight == 0 {
			cc.weight = 10 // 默认权重
		}

		cc.currentWeight = cc.weight
		conns = append(conns, cc)
	}
	return &wrrPicker{conns: conns}
}

type wrrPicker struct {
	conns []*conn
	mutex sync.Mutex
}

// Pick 实现平滑加权负载均衡算法
func (p *wrrPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 节点为空
	if len(p.conns) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	var total int
	var maxCC *conn
	for _, cc := range p.conns {
		total += cc.weight
		cc.currentWeight = cc.currentWeight + cc.weight
		// 找到当前权重最大的节点
		if maxCC == nil || cc.currentWeight > maxCC.currentWeight {
			maxCC = cc
		}
	}
	// maxCC 就是选中的节点，
	// 节点被选中后，更新权重，避免节点下一轮选中
	maxCC.currentWeight = maxCC.currentWeight - total

	return balancer.PickResult{
		SubConn: maxCC.cc,
		Done: func(info balancer.DoneInfo) {
			// 很多动态算法，就是在这里做一些事情处理，比如根据结果调整权重

			// 节点被选中后，更新权重
			// maxCC.currentWeight = maxCC.currentWeight - maxCC.weight
		},
	}, nil
}

// conn 表示节点
type conn struct {
	weight        int              // 权重
	currentWeight int              // 当前权重
	cc            balancer.SubConn // 节点，在gRPC中代表真正的一个节点
}
