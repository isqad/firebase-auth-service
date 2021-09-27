package service

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor is interface for declare auth interceptor
type AuthInterceptor interface {
	GetFirebaseAuthServiceAddr() string

	AuthUnaryInterceptor(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error)

	AuthStreamInterceptor(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error
}

type interceptor struct {
	addr string
}

// NewAuthInterceptor returns new instance of auth interceptor
func NewAuthInterceptor(authServiceAddress string) AuthInterceptor {
	return &interceptor{
		addr: authServiceAddress,
	}
}

func (i *interceptor) GetFirebaseAuthServiceAddr() string {
	return i.addr
}

// AuthUnaryInterceptor send jwt to auth service before handle unary request
func (i *interceptor) AuthUnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	err := i.authenticate(ctx, info.FullMethod)
	if err != nil {
		return nil, err
	}

	return handler(ctx, req)
}

// AuthStreamInterceptor returns stream interceptor for authenticate
func (i *interceptor) AuthStreamInterceptor(
	srv interface{},
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	err := i.authenticate(stream.Context(), info.FullMethod)
	if err != nil {
		return err
	}

	return handler(srv, stream)
}

func (i *interceptor) authenticate(ctx context.Context, method string) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md["authorization"]) == 0 {
		return status.Errorf(codes.Unauthenticated, "Wrong metadata: %v", md)
	}

	conn, err := grpc.Dial(i.GetFirebaseAuthServiceAddr(), []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
	}...)
	if err != nil {
		return status.Errorf(codes.Internal, "Could not to dial auth server")
	}
	defer conn.Close()
	authClient := NewAuthClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = authClient.Verify(ctx, &Token{Token: md["authorization"][0]})
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "Could not to verify auth token")
	}

	return nil
}
