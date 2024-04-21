package scratch

import (
	"context"
	"errors"
	"fmt"
	"github.com/Yakwilik/MRGAbackend/internal/logger"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"sync"
)

type Service interface {
	GetDescription() ServiceDesc
}

type ServiceDesc interface {
	RegisterGRPC(server *grpc.Server)
	RegisterGateway(ctx context.Context, mux *runtime.ServeMux) error
}

type App struct {
	desc       ServiceDesc
	mux        *http.ServeMux
	gatewayMux *runtime.ServeMux
	grpcServer *grpc.Server
	opts       *Options
	lis        *listeners
	wg         *sync.WaitGroup
}

var mdOption = runtime.WithMetadata(func(ctx context.Context, request *http.Request) metadata.MD {
	outMD := metadata.MD{}

	for key, val := range request.Header {
		if strings.EqualFold(key, "Cookie") {
			outMD[key] = val
		}
	}
	return outMD
})

var headerOption = runtime.WithOutgoingHeaderMatcher(func(s string) (string, bool) {
	if strings.EqualFold(s, "Set-Cookie") {
		return s, true
	}
	return s, false
})

func defaultOptions() []Option {
	options := make([]Option, 0, 2)

	options = append(options, optionFn(func(options *Options) error {
		if options != nil {
			options.ServeMuxOpts = append(options.ServeMuxOpts,
				mdOption, headerOption)
		}
		return nil
	}))
	return options
}
func InitApp(opts ...Option) (*App, error) {
	o, err := evaluateOptions(append(opts, defaultOptions()...))
	if err != nil {
		return nil, err
	}

	logOptions := logger.Options{
		OutputAddr:  o.LoggerOutput,
		Environment: o.Environment,
		AppName:     o.AppName,
		LogLevel:    o.LogLevel,
	}
	logger.InitLogger(logOptions)

	a := &App{
		opts: o,
		wg:   &sync.WaitGroup{},
	}

	lst, err := newListeners(a.opts)
	if err != nil {
		return nil, fmt.Errorf("can't start listeners: %w", err)
	}
	a.lis = lst

	return a, nil
}

func (a *App) Run(impl ...Service) error {
	descs := make([]ServiceDesc, 0, len(impl))
	for _, i := range impl {
		descs = append(descs, i.GetDescription())
	}

	a.desc = NewCompoundServiceDesc(descs...)

	a.initPublicHTTP()
	a.initGRPCServer(NewCompoundServiceDesc(a.desc))
	a.initPublicHTTPHandlers(a.desc)
	a.runPublicHTTP()
	a.runGRPC()
	a.wg.Wait()
	return nil
}
func (a *App) runGRPC() {
	a.wg.Add(1)

	if a.grpcServer != nil {
		log.Println("running grpc")
		go func() {
			defer a.wg.Done()
			if err := a.grpcServer.Serve(a.lis.grpc); err != nil {
				log.Print(fmt.Errorf("grpc.public: %w", err))
			}
		}()
	}
}
func (a *App) runPublicHTTP() {
	a.wg.Add(1)

	var h http.Handler = a.mux
	if a.opts.EnablePublicMuxMiddleware {
		for _, mw := range a.opts.PublicMuxMiddleware {
			slog.Debug("adding public mux middleware")
			h = mw(h)
		}
	}
	publicServer := &http.Server{
		Handler: h,
	}

	go func() {
		defer a.wg.Done()
		if err := publicServer.Serve(a.lis.http); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Print(fmt.Errorf("http.public: %w", err))
		}
	}()
}

func (a *App) initGRPCServer(desc ServiceDesc) {
	if desc == nil {
		return
	}

	slog.Info("initializing gRPC server", "interceptor", "LogInterceptor")
	a.grpcServer = grpc.NewServer(grpc.UnaryInterceptor(LogInterceptor))
	desc.RegisterGRPC(a.grpcServer)
	reflection.Register(a.grpcServer)
}

func (a *App) initPublicHTTPHandlers(desc ServiceDesc) {
	if desc != nil {
		if err := desc.RegisterGateway(context.Background(), a.gatewayMux); err != nil {
			log.Fatalf("error while register gateway: %v", err)
		}
	}
}

func (a *App) initPublicHTTP() {
	a.gatewayMux = runtime.NewServeMux(a.opts.ServeMuxOpts...)
	a.mux = http.NewServeMux()
	if a.opts.EnableCustomHandler {
		h := a.opts.CustomHandler
		if a.opts.EnableCustomMuxMiddleware {
			for _, mw := range a.opts.CustomMuxMiddleware {
				if mw != nil {
					h = mw(h)
				}
			}
		}
		a.mux.Handle("/api/", http.StripPrefix("/api", h))
	}

	var gatewayMux http.Handler = a.gatewayMux
	if a.opts.EnableGatewayMiddleware {
		for _, mw := range a.opts.GatewayMiddleware {
			if mw != nil {
				gatewayMux = mw(gatewayMux)
			}
		}
	}
	a.mux.Handle("/api/gateway/", http.StripPrefix("/api/gateway", gatewayMux))
}
