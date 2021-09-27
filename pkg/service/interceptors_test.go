package service

import (
	"context"
	"testing"

	"google.golang.org/grpc"
)

func unaryHandler(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

func TestAuthUnaryInterceptor(t *testing.T) {
	interceptor := NewAuthInterceptor("127.0.0.1:50053")
	srvInfo := &grpc.UnaryServerInfo{FullMethod: "/service.auth"}
	_, err := interceptor.AuthUnaryInterceptor(context.Background(), nil, srvInfo, unaryHandler)
	if err == nil {
		t.Errorf("Expected err not nil, got: %v", err)
	}
}
