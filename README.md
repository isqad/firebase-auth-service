<h1 align="center">Grpc-service for Firebase JWTs authentication</h1>

[![Go](https://github.com/isqad/firebase-auth-service/actions/workflows/go.yml/badge.svg)](https://github.com/isqad/firebase-auth-service/actions/workflows/go.yml)

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

2) Run the service:

```bash
docker run -d -p 50053:50053 --name firebase-auth-service isqad88/firebase-auth-service:0.1.0
```

3) Run your grpc-service.
4) Verify your JWTs from firebase auth:

```go
// Example

import (
  // Some imports...
  firebase "github.com/isqad/firebase-auth-service/pkg/service"
  // Other imports...
)

func someHandler(w http.ResponseWriter, r *http.Request) {
  // Extract token from request:
  token := r.Header.Get("X-Auth-Token")

  conn, err := grpc.Dial(m.Addr, []grpc.DialOption{
      grpc.WithInsecure(),
      grpc.WithBlock(),
  }...)
  if err != nil {
      authFailed(w, r, err)
      return
  }
  defer conn.Close()

  authClient := firebase.NewAuthClient(conn)
  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
  defer cancel()

  t, err := authClient.Verify(ctx, &firebase.Token{Token: token})
  if err != nil {
      m.authFailed(w, r, err)
      return
  }
```

# Roadmap

- [x] Grabing google pub keys
- [x] Verification firebase auth JWT of User and extract his ID
- [x] grpc auth interceptors for services
- [x] Configuration
- [ ] Tests
- [x] grpc error codes
