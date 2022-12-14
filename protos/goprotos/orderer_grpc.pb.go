// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package protos

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// OrdererServiceClient is the client API for OrdererService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type OrdererServiceClient interface {
	CommitFabricAndFabricCRDTTransactionStream(ctx context.Context, opts ...grpc.CallOption) (OrdererService_CommitFabricAndFabricCRDTTransactionStreamClient, error)
	StopAndGetProfilingResult(ctx context.Context, in *Profiling, opts ...grpc.CallOption) (OrdererService_StopAndGetProfilingResultClient, error)
	SubscribeBlocks(ctx context.Context, in *BlockEventSubscription, opts ...grpc.CallOption) (OrdererService_SubscribeBlocksClient, error)
	ChangeModeRestart(ctx context.Context, in *OperationMode, opts ...grpc.CallOption) (*Empty, error)
}

type ordererServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewOrdererServiceClient(cc grpc.ClientConnInterface) OrdererServiceClient {
	return &ordererServiceClient{cc}
}

func (c *ordererServiceClient) CommitFabricAndFabricCRDTTransactionStream(ctx context.Context, opts ...grpc.CallOption) (OrdererService_CommitFabricAndFabricCRDTTransactionStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &OrdererService_ServiceDesc.Streams[0], "/protos.OrdererService/CommitFabricAndFabricCRDTTransactionStream", opts...)
	if err != nil {
		return nil, err
	}
	x := &ordererServiceCommitFabricAndFabricCRDTTransactionStreamClient{stream}
	return x, nil
}

type OrdererService_CommitFabricAndFabricCRDTTransactionStreamClient interface {
	Send(*Transaction) error
	CloseAndRecv() (*Empty, error)
	grpc.ClientStream
}

type ordererServiceCommitFabricAndFabricCRDTTransactionStreamClient struct {
	grpc.ClientStream
}

func (x *ordererServiceCommitFabricAndFabricCRDTTransactionStreamClient) Send(m *Transaction) error {
	return x.ClientStream.SendMsg(m)
}

func (x *ordererServiceCommitFabricAndFabricCRDTTransactionStreamClient) CloseAndRecv() (*Empty, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(Empty)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *ordererServiceClient) StopAndGetProfilingResult(ctx context.Context, in *Profiling, opts ...grpc.CallOption) (OrdererService_StopAndGetProfilingResultClient, error) {
	stream, err := c.cc.NewStream(ctx, &OrdererService_ServiceDesc.Streams[1], "/protos.OrdererService/StopAndGetProfilingResult", opts...)
	if err != nil {
		return nil, err
	}
	x := &ordererServiceStopAndGetProfilingResultClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type OrdererService_StopAndGetProfilingResultClient interface {
	Recv() (*ProfilingResult, error)
	grpc.ClientStream
}

type ordererServiceStopAndGetProfilingResultClient struct {
	grpc.ClientStream
}

func (x *ordererServiceStopAndGetProfilingResultClient) Recv() (*ProfilingResult, error) {
	m := new(ProfilingResult)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *ordererServiceClient) SubscribeBlocks(ctx context.Context, in *BlockEventSubscription, opts ...grpc.CallOption) (OrdererService_SubscribeBlocksClient, error) {
	stream, err := c.cc.NewStream(ctx, &OrdererService_ServiceDesc.Streams[2], "/protos.OrdererService/SubscribeBlocks", opts...)
	if err != nil {
		return nil, err
	}
	x := &ordererServiceSubscribeBlocksClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type OrdererService_SubscribeBlocksClient interface {
	Recv() (*Block, error)
	grpc.ClientStream
}

type ordererServiceSubscribeBlocksClient struct {
	grpc.ClientStream
}

func (x *ordererServiceSubscribeBlocksClient) Recv() (*Block, error) {
	m := new(Block)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *ordererServiceClient) ChangeModeRestart(ctx context.Context, in *OperationMode, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/protos.OrdererService/ChangeModeRestart", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// OrdererServiceServer is the server API for OrdererService service.
// All implementations should embed UnimplementedOrdererServiceServer
// for forward compatibility
type OrdererServiceServer interface {
	CommitFabricAndFabricCRDTTransactionStream(OrdererService_CommitFabricAndFabricCRDTTransactionStreamServer) error
	StopAndGetProfilingResult(*Profiling, OrdererService_StopAndGetProfilingResultServer) error
	SubscribeBlocks(*BlockEventSubscription, OrdererService_SubscribeBlocksServer) error
	ChangeModeRestart(context.Context, *OperationMode) (*Empty, error)
}

// UnimplementedOrdererServiceServer should be embedded to have forward compatible implementations.
type UnimplementedOrdererServiceServer struct {
}

func (UnimplementedOrdererServiceServer) CommitFabricAndFabricCRDTTransactionStream(OrdererService_CommitFabricAndFabricCRDTTransactionStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method CommitFabricAndFabricCRDTTransactionStream not implemented")
}
func (UnimplementedOrdererServiceServer) StopAndGetProfilingResult(*Profiling, OrdererService_StopAndGetProfilingResultServer) error {
	return status.Errorf(codes.Unimplemented, "method StopAndGetProfilingResult not implemented")
}
func (UnimplementedOrdererServiceServer) SubscribeBlocks(*BlockEventSubscription, OrdererService_SubscribeBlocksServer) error {
	return status.Errorf(codes.Unimplemented, "method SubscribeBlocks not implemented")
}
func (UnimplementedOrdererServiceServer) ChangeModeRestart(context.Context, *OperationMode) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChangeModeRestart not implemented")
}

// UnsafeOrdererServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to OrdererServiceServer will
// result in compilation errors.
type UnsafeOrdererServiceServer interface {
	mustEmbedUnimplementedOrdererServiceServer()
}

func RegisterOrdererServiceServer(s grpc.ServiceRegistrar, srv OrdererServiceServer) {
	s.RegisterService(&OrdererService_ServiceDesc, srv)
}

func _OrdererService_CommitFabricAndFabricCRDTTransactionStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(OrdererServiceServer).CommitFabricAndFabricCRDTTransactionStream(&ordererServiceCommitFabricAndFabricCRDTTransactionStreamServer{stream})
}

type OrdererService_CommitFabricAndFabricCRDTTransactionStreamServer interface {
	SendAndClose(*Empty) error
	Recv() (*Transaction, error)
	grpc.ServerStream
}

type ordererServiceCommitFabricAndFabricCRDTTransactionStreamServer struct {
	grpc.ServerStream
}

func (x *ordererServiceCommitFabricAndFabricCRDTTransactionStreamServer) SendAndClose(m *Empty) error {
	return x.ServerStream.SendMsg(m)
}

func (x *ordererServiceCommitFabricAndFabricCRDTTransactionStreamServer) Recv() (*Transaction, error) {
	m := new(Transaction)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _OrdererService_StopAndGetProfilingResult_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Profiling)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(OrdererServiceServer).StopAndGetProfilingResult(m, &ordererServiceStopAndGetProfilingResultServer{stream})
}

type OrdererService_StopAndGetProfilingResultServer interface {
	Send(*ProfilingResult) error
	grpc.ServerStream
}

type ordererServiceStopAndGetProfilingResultServer struct {
	grpc.ServerStream
}

func (x *ordererServiceStopAndGetProfilingResultServer) Send(m *ProfilingResult) error {
	return x.ServerStream.SendMsg(m)
}

func _OrdererService_SubscribeBlocks_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(BlockEventSubscription)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(OrdererServiceServer).SubscribeBlocks(m, &ordererServiceSubscribeBlocksServer{stream})
}

type OrdererService_SubscribeBlocksServer interface {
	Send(*Block) error
	grpc.ServerStream
}

type ordererServiceSubscribeBlocksServer struct {
	grpc.ServerStream
}

func (x *ordererServiceSubscribeBlocksServer) Send(m *Block) error {
	return x.ServerStream.SendMsg(m)
}

func _OrdererService_ChangeModeRestart_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OperationMode)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrdererServiceServer).ChangeModeRestart(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.OrdererService/ChangeModeRestart",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrdererServiceServer).ChangeModeRestart(ctx, req.(*OperationMode))
	}
	return interceptor(ctx, in, info, handler)
}

// OrdererService_ServiceDesc is the grpc.ServiceDesc for OrdererService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var OrdererService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "protos.OrdererService",
	HandlerType: (*OrdererServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ChangeModeRestart",
			Handler:    _OrdererService_ChangeModeRestart_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "CommitFabricAndFabricCRDTTransactionStream",
			Handler:       _OrdererService_CommitFabricAndFabricCRDTTransactionStream_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "StopAndGetProfilingResult",
			Handler:       _OrdererService_StopAndGetProfilingResult_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "SubscribeBlocks",
			Handler:       _OrdererService_SubscribeBlocks_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "orderer.proto",
}
