<h1 align="center">Grpc-service for Firebase JWTs authentication</h1>

Grpc-service for verification and authentication of Firebase JWTs

# Usage

1) Integrate auth interceptors into your grpc-project:

```go
package main

import (
  "net"
  "google.golang.org/grpc"
  auth "github.com/isqad/firebase-auth-service/pkg/service"
)

type UsersAPIServer struct {
	UnimplementedUsersServer
}

func NewUsersAPIServer() (*UsersAPIServer, error) {
  return &UsersAPIServer{}
}

func (s *UsersAPIServer) Start() error {
  authAddr := os.Getenv("FIREBASE_AUTH_SERVICE_ADDRESS")

  l, err := net.Listen("tcp", authAddr)
	if err != nil {
		return err
	}

  a := auth.NewAuthInterceptor(authAddr)

	g := grpc.NewServer(
		grpc.UnaryInterceptor(a.AuthUnaryInterceptor),
    grpc.StreamInterceptor(a.AuthStreamInterceptor),
	)
	RegisterUsersServer(g, s)
	if err := g.Serve(l); err != nil {
		return err
	}

	return nil
}
```

2) Build and run the service:

```bash
docker build -t your-name:firebase-auth-service:0.1.0 .
```

```bash
docker run --rm -d -p 50053:50053 --name firebase-auth-service your-name:firebase-auth-service your-name:firebase-auth-service:0.1.0
```

3) Run your grpc-service.

# Roadmap

- [x] Grabing google pub keys
- [x] Verification firebase auth JWT of User and extract his ID
- [x] grpc auth interceptors for services
- [x] Configuration
- [ ] Tests
- [x] grpc error codes
