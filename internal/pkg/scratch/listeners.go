package scratch

import (
	"fmt"
	"net"
)

type listeners struct {
	http net.Listener
	grpc net.Listener
}

func newListeners(opt *Options) (*listeners, error) {
	host := opt.BindAddress

	l := &listeners{}

	http, err := listenPublicHTTP(opt)
	if err != nil {
		return nil, fmt.Errorf("couldn't create listener: %w", err)
	}

	l.http = http

	grpc, err := net.Listen("tcp", fmt.Sprintf("%s:%v", host, opt.PortGRPC))
	if err != nil {
		return nil, fmt.Errorf("couldn't create RPC listener: %w", err)
	}
	l.grpc = grpc

	return l, nil
}

func listenPublicHTTP(opt *Options) (net.Listener, error) {
	listenAddr := fmt.Sprintf("%s:%v", opt.BindAddress, opt.PortHTTP)

	http, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, fmt.Errorf("couldn't create HTTP listener: %w", err)
	}

	return http, nil

}
