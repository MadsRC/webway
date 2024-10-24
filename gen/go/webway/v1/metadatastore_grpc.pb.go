// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.4.0
// - protoc             (unknown)
// source: webway/v1/metadatastore.proto

package v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.62.0 or later.
const _ = grpc.SupportPackageIsVersion8

const (
	MetadataStore_RegisterAgent_FullMethodName   = "/webway.v1.MetadataStore/RegisterAgent"
	MetadataStore_DeregisterAgent_FullMethodName = "/webway.v1.MetadataStore/DeregisterAgent"
	MetadataStore_GetMetadata_FullMethodName     = "/webway.v1.MetadataStore/GetMetadata"
	MetadataStore_AgentHeartbeat_FullMethodName  = "/webway.v1.MetadataStore/AgentHeartbeat"
)

// MetadataStoreClient is the client API for MetadataStore service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MetadataStoreClient interface {
	RegisterAgent(ctx context.Context, in *RegisterAgentRequest, opts ...grpc.CallOption) (*RegisterAgentResponse, error)
	DeregisterAgent(ctx context.Context, in *DeregisterAgentRequest, opts ...grpc.CallOption) (*DeregisterAgentResponse, error)
	GetMetadata(ctx context.Context, in *GetMetadataRequest, opts ...grpc.CallOption) (*GetMetadataResponse, error)
	AgentHeartbeat(ctx context.Context, in *Heartbeat, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type metadataStoreClient struct {
	cc grpc.ClientConnInterface
}

func NewMetadataStoreClient(cc grpc.ClientConnInterface) MetadataStoreClient {
	return &metadataStoreClient{cc}
}

func (c *metadataStoreClient) RegisterAgent(ctx context.Context, in *RegisterAgentRequest, opts ...grpc.CallOption) (*RegisterAgentResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RegisterAgentResponse)
	err := c.cc.Invoke(ctx, MetadataStore_RegisterAgent_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metadataStoreClient) DeregisterAgent(ctx context.Context, in *DeregisterAgentRequest, opts ...grpc.CallOption) (*DeregisterAgentResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DeregisterAgentResponse)
	err := c.cc.Invoke(ctx, MetadataStore_DeregisterAgent_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metadataStoreClient) GetMetadata(ctx context.Context, in *GetMetadataRequest, opts ...grpc.CallOption) (*GetMetadataResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetMetadataResponse)
	err := c.cc.Invoke(ctx, MetadataStore_GetMetadata_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metadataStoreClient) AgentHeartbeat(ctx context.Context, in *Heartbeat, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, MetadataStore_AgentHeartbeat_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MetadataStoreServer is the server API for MetadataStore service.
// All implementations must embed UnimplementedMetadataStoreServer
// for forward compatibility
type MetadataStoreServer interface {
	RegisterAgent(context.Context, *RegisterAgentRequest) (*RegisterAgentResponse, error)
	DeregisterAgent(context.Context, *DeregisterAgentRequest) (*DeregisterAgentResponse, error)
	GetMetadata(context.Context, *GetMetadataRequest) (*GetMetadataResponse, error)
	AgentHeartbeat(context.Context, *Heartbeat) (*emptypb.Empty, error)
	mustEmbedUnimplementedMetadataStoreServer()
}

// UnimplementedMetadataStoreServer must be embedded to have forward compatible implementations.
type UnimplementedMetadataStoreServer struct {
}

func (UnimplementedMetadataStoreServer) RegisterAgent(context.Context, *RegisterAgentRequest) (*RegisterAgentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterAgent not implemented")
}
func (UnimplementedMetadataStoreServer) DeregisterAgent(context.Context, *DeregisterAgentRequest) (*DeregisterAgentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeregisterAgent not implemented")
}
func (UnimplementedMetadataStoreServer) GetMetadata(context.Context, *GetMetadataRequest) (*GetMetadataResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMetadata not implemented")
}
func (UnimplementedMetadataStoreServer) AgentHeartbeat(context.Context, *Heartbeat) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AgentHeartbeat not implemented")
}
func (UnimplementedMetadataStoreServer) mustEmbedUnimplementedMetadataStoreServer() {}

// UnsafeMetadataStoreServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MetadataStoreServer will
// result in compilation errors.
type UnsafeMetadataStoreServer interface {
	mustEmbedUnimplementedMetadataStoreServer()
}

func RegisterMetadataStoreServer(s grpc.ServiceRegistrar, srv MetadataStoreServer) {
	s.RegisterService(&MetadataStore_ServiceDesc, srv)
}

func _MetadataStore_RegisterAgent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterAgentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetadataStoreServer).RegisterAgent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MetadataStore_RegisterAgent_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetadataStoreServer).RegisterAgent(ctx, req.(*RegisterAgentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetadataStore_DeregisterAgent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeregisterAgentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetadataStoreServer).DeregisterAgent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MetadataStore_DeregisterAgent_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetadataStoreServer).DeregisterAgent(ctx, req.(*DeregisterAgentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetadataStore_GetMetadata_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetMetadataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetadataStoreServer).GetMetadata(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MetadataStore_GetMetadata_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetadataStoreServer).GetMetadata(ctx, req.(*GetMetadataRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetadataStore_AgentHeartbeat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Heartbeat)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetadataStoreServer).AgentHeartbeat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MetadataStore_AgentHeartbeat_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetadataStoreServer).AgentHeartbeat(ctx, req.(*Heartbeat))
	}
	return interceptor(ctx, in, info, handler)
}

// MetadataStore_ServiceDesc is the grpc.ServiceDesc for MetadataStore service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MetadataStore_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "webway.v1.MetadataStore",
	HandlerType: (*MetadataStoreServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RegisterAgent",
			Handler:    _MetadataStore_RegisterAgent_Handler,
		},
		{
			MethodName: "DeregisterAgent",
			Handler:    _MetadataStore_DeregisterAgent_Handler,
		},
		{
			MethodName: "GetMetadata",
			Handler:    _MetadataStore_GetMetadata_Handler,
		},
		{
			MethodName: "AgentHeartbeat",
			Handler:    _MetadataStore_AgentHeartbeat_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "webway/v1/metadatastore.proto",
}
