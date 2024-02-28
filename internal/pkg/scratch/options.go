package scratch

import (
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

var (
	PublicPort  uint = 7001
	GrpcPort    uint = 7002
	BindAddress      = ""
)

type Options struct {
	PortHTTP uint
	PortGRPC uint

	BindAddress string

	ServeMuxOpts []runtime.ServeMuxOption
}

type Option interface {
	Apply(options *Options) error
}

type optionFn func(options *Options) error

func (fn optionFn) Apply(opts *Options) error {
	return fn(opts)
}

func evaluateOptions(opts []Option) (*Options, error) {
	oo := &Options{
		PortHTTP:    PublicPort,
		PortGRPC:    GrpcPort,
		BindAddress: BindAddress,
	}

	for _, o := range opts {
		if err := o.Apply(oo); err != nil {
			return nil, fmt.Errorf("invalid option: %w", err)
		}
	}

	return oo, nil
}
