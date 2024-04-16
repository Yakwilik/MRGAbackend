package scratch

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"log"
)

type CompoundServiceDesc struct {
	svc []ServiceDesc
}

func NewCompoundServiceDesc(desc ...ServiceDesc) *CompoundServiceDesc {
	return &CompoundServiceDesc{svc: desc}
}

func (d *CompoundServiceDesc) RegisterGRPC(g *grpc.Server) {
	for _, svc := range d.svc {
		svc.RegisterGRPC(g)
	}
}

func (d *CompoundServiceDesc) RegisterGateway(ctx context.Context, mux *runtime.ServeMux) error {
	for _, svc := range d.svc {
		if err := svc.RegisterGateway(ctx, mux); err != nil {
			log.Print(err)
		}
	}

	return nil
}
