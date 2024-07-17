package grpc

import (
	"github.com/madsrc/webway"
	pb "github.com/madsrc/webway/gen/go/webway/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

type Server struct {
	opts       serverOptions
	grpcServer *grpc.Server
}

func NewServer(opt ...ServerOption) (*Server, error) {
	opts := defaultServerOptions
	for _, o := range globalServerOptions {
		o.apply(&opts)
	}
	for _, o := range opt {
		o.apply(&opts)
	}
	s := &Server{opts: opts}

	s.grpcServer = grpc.NewServer()

	if s.opts.MetadataStoreServer == nil {
		return nil, &webway.MissingOptionError{Option: "MetadataStoreServer"}
	}

	pb.RegisterMetadataStoreServer(s.grpcServer, s.opts.MetadataStoreServer)

	if s.opts.ReflectionEnabled {
		reflection.Register(s.grpcServer)
	}

	return s, nil
}

func (s *Server) GracefulStop() {
	s.grpcServer.GracefulStop()
}

func (s *Server) Serve(lis net.Listener) error {
	return s.grpcServer.Serve(lis)
}

type serverOptions struct {
	MetadataStoreServer pb.MetadataStoreServer
	ReflectionEnabled   bool
}

var defaultServerOptions = serverOptions{
	ReflectionEnabled: true,
}
var globalServerOptions []ServerOption

type ServerOption interface {
	apply(*serverOptions)
}

// funcServerOption wraps a function that modifies serverOptions into an
// implementation of the ServerOption interface.
type funcServerOption struct {
	f func(*serverOptions)
}

func (fdo *funcServerOption) apply(do *serverOptions) {
	fdo.f(do)
}

func newFuncServerOption(f func(*serverOptions)) *funcServerOption {
	return &funcServerOption{
		f: f,
	}
}

func WithServerMetadataStoreServer(mds pb.MetadataStoreServer) ServerOption {
	return newFuncServerOption(func(o *serverOptions) {
		o.MetadataStoreServer = mds
	})
}

func WithServerReflection() ServerOption {
	return newFuncServerOption(func(o *serverOptions) {
		o.ReflectionEnabled = true
	})
}

func WithoutServerReflection() ServerOption {
	return newFuncServerOption(func(o *serverOptions) {
		o.ReflectionEnabled = false
	})
}
