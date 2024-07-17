package grpc

import (
	"context"
	"google.golang.org/grpc"
)

type ClientConn struct {
	opts           clientConnOptions
	grpcClientConn *grpc.ClientConn
}

func NewClientConn(target string, opt ...ClientConnOption) (*ClientConn, error) {
	opts := defaultClientConnOptions
	for _, o := range globalClientConnOptions {
		o.apply(&opts)
	}
	for _, o := range opt {
		o.apply(&opts)
	}
	s := &ClientConn{opts: opts}

	grpcDialOptions := make([]grpc.DialOption, 0)

	if opts.Insecure {
		grpcDialOptions = append(grpcDialOptions, grpc.WithInsecure())
	}

	var err error
	s.grpcClientConn, err = grpc.NewClient(target, grpcDialOptions...)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *ClientConn) GracefulStop() {
	_ = s.grpcClientConn.Close()
}

func (s *ClientConn) Invoke(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error {
	return s.grpcClientConn.Invoke(ctx, method, args, reply, opts...)
}

func (s *ClientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return s.grpcClientConn.NewStream(ctx, desc, method, opts...)
}

type clientConnOptions struct {
	Insecure bool
}

var defaultClientConnOptions = clientConnOptions{
	Insecure: false,
}
var globalClientConnOptions []ClientConnOption

type ClientConnOption interface {
	apply(*clientConnOptions)
}

// funcClientConnOption wraps a function that modifies clientConnOptions into an
// implementation of the ClientConnOption interface.
type funcClientConnOption struct {
	f func(*clientConnOptions)
}

func (fdo *funcClientConnOption) apply(do *clientConnOptions) {
	fdo.f(do)
}

func newFuncClientConnOption(f func(*clientConnOptions)) *funcClientConnOption {
	return &funcClientConnOption{
		f: f,
	}
}

func WithClientConnInsecure() ClientConnOption {
	return newFuncClientConnOption(func(o *clientConnOptions) {
		o.Insecure = true
	})
}
