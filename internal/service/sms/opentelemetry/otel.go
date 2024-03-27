package opentelemetry

import (
	"context"

	"github.com/mrhelloboy/wehook/internal/service/sms"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Service struct {
	svc    sms.Service
	tracer trace.Tracer
}

func (s *Service) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	ctx, span := s.tracer.Start(ctx, "sms_send_"+biz, trace.WithSpanKind(trace.SpanKindClient))
	defer span.End(trace.WithStackTrace(true))
	err := s.svc.Send(ctx, biz, args, numbers...)
	if err != nil {
		span.RecordError(err)
	}
	return err
}

func NewService(svc sms.Service) *Service {
	tp := otel.GetTracerProvider()
	tracer := tp.Tracer("github.com/mrhelloboy/wehook/internal/service/sms/opentelemetry")
	return &Service{
		svc:    svc,
		tracer: tracer,
	}
}
