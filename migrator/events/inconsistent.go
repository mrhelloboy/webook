package events

import "context"

type Producer interface {
	ProduceInconsistentEvent(ctx context.Context, evt InconsistentEvent) error
}

type InconsistentEvent struct {
	ID int64

	// 值为 SRC 或 DST
	// SRC：以源表为准
	// DST：以目标表为准
	Direction string
	Type      string
}

const (
	// InconsistentEventTypeTargetMissing 校验的目标数据，缺了这一条
	InconsistentEventTypeTargetMissing = "target_missing"
	// InconsistentEventTypeNEQ 不相等
	InconsistentEventTypeNEQ         = "neq"
	InconsistentEventTypeBaseMissing = "base_missing"
)
