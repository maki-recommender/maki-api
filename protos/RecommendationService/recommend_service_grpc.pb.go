// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.24.0
// source: protos/recommend_service.proto

package RecommendationService

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

const (
	RecommendationService_GetAnimeRecommendations_FullMethodName = "/maki.RecommendationService/GetAnimeRecommendations"
)

// RecommendationServiceClient is the client API for RecommendationService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RecommendationServiceClient interface {
	// Get a list of reccomendations based on the user list given as input
	GetAnimeRecommendations(ctx context.Context, in *WatchedAnime, opts ...grpc.CallOption) (*RecommendedAnime, error)
}

type recommendationServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewRecommendationServiceClient(cc grpc.ClientConnInterface) RecommendationServiceClient {
	return &recommendationServiceClient{cc}
}

func (c *recommendationServiceClient) GetAnimeRecommendations(ctx context.Context, in *WatchedAnime, opts ...grpc.CallOption) (*RecommendedAnime, error) {
	out := new(RecommendedAnime)
	err := c.cc.Invoke(ctx, RecommendationService_GetAnimeRecommendations_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RecommendationServiceServer is the server API for RecommendationService service.
// All implementations must embed UnimplementedRecommendationServiceServer
// for forward compatibility
type RecommendationServiceServer interface {
	// Get a list of reccomendations based on the user list given as input
	GetAnimeRecommendations(context.Context, *WatchedAnime) (*RecommendedAnime, error)
	mustEmbedUnimplementedRecommendationServiceServer()
}

// UnimplementedRecommendationServiceServer must be embedded to have forward compatible implementations.
type UnimplementedRecommendationServiceServer struct {
}

func (UnimplementedRecommendationServiceServer) GetAnimeRecommendations(context.Context, *WatchedAnime) (*RecommendedAnime, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAnimeRecommendations not implemented")
}
func (UnimplementedRecommendationServiceServer) mustEmbedUnimplementedRecommendationServiceServer() {}

// UnsafeRecommendationServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RecommendationServiceServer will
// result in compilation errors.
type UnsafeRecommendationServiceServer interface {
	mustEmbedUnimplementedRecommendationServiceServer()
}

func RegisterRecommendationServiceServer(s grpc.ServiceRegistrar, srv RecommendationServiceServer) {
	s.RegisterService(&RecommendationService_ServiceDesc, srv)
}

func _RecommendationService_GetAnimeRecommendations_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WatchedAnime)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RecommendationServiceServer).GetAnimeRecommendations(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RecommendationService_GetAnimeRecommendations_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RecommendationServiceServer).GetAnimeRecommendations(ctx, req.(*WatchedAnime))
	}
	return interceptor(ctx, in, info, handler)
}

// RecommendationService_ServiceDesc is the grpc.ServiceDesc for RecommendationService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RecommendationService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "maki.RecommendationService",
	HandlerType: (*RecommendationServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetAnimeRecommendations",
			Handler:    _RecommendationService_GetAnimeRecommendations_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "protos/recommend_service.proto",
}
