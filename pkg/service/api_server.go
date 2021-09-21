package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	log "github.com/sirupsen/logrus"
)

const pubKeyURL = "https://www.googleapis.com/robot/v1/metadata/x509/securetoken@system.gserviceaccount.com"

var cacheControlMaxAgeRe = regexp.MustCompile(`max-age=(\d+)`)

type pubKeys struct {
	keys        map[string]string
	m           sync.Mutex
	client      *http.Client
	maxAge      int
	refreshedAt time.Time
}

type firebaseJWTClaims struct {
	AuthTime int `json:"auth_time,omitempty"`
	jwt.StandardClaims
}
type APIServer struct {
	UnimplementedAuthServer
	keys *pubKeys
}

type ConfigKey string

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

// NewAPIServer builds new instance of the ApiServer
func NewAPIServer() (*APIServer, error) {
	keys, err := grabPubKeys()
	if err != nil {
		return nil, err
	}
	return &APIServer{keys: keys}, nil
}

func (s *APIServer) Verify(ctx context.Context, token *Token) (*Token, error) {
	log.WithFields(log.Fields{
		"token": token,
	}).Info("Got verify request")

	t, err := jwt.ParseWithClaims(token.Token, &firebaseJWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, status.Error(codes.Unauthenticated, fmt.Sprintf("Unexpected signing method: %v", t.Header["alg"]))
		}
		untypedKeyID, ok := t.Header["kid"]
		if !ok {
			return nil, status.Error(codes.Unauthenticated, fmt.Sprintf("No key ID ing header, header: %v", untypedKeyID))
		}

		keyID := untypedKeyID.(string)
		keyStr, err := s.keys.get(keyID)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, fmt.Sprintf("Can not get pub key: %v", err))
		}
		pubKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(keyStr))
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, fmt.Sprintf("An error occurred parsing the public key base64 for key ID '%v'; this is a code bug", untypedKeyID))
		}

		return pubKey, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := t.Claims.(*firebaseJWTClaims); ok && t.Valid {
		token.UserId = &claims.Subject

		log.WithFields(log.Fields{
			"token": token,
		}).Info("Successfully verified")

		return token, nil
	}
	return nil, status.Error(codes.Unauthenticated, "Invalid token")
}

func (s *APIServer) Start(ctx *cli.Context) error {
	log.Info("Starting server...")
	log.WithFields(log.Fields{"listen-address": ctx.String("listen-address")}).Info("Given listen address")

	l, err := net.Listen("tcp", ctx.String("listen-address"))
	if err != nil {
		return err
	}

	g := grpc.NewServer()
	RegisterAuthServer(g, s)
	if err := g.Serve(l); err != nil {
		return err
	}

	return nil
}

func grabPubKeys() (*pubKeys, error) {
	k := &pubKeys{
		m:    sync.Mutex{},
		keys: make(map[string]string),
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
	err := k.refresh()
	if err != nil {
		return nil, err
	}

	return k, nil
}

func (k *pubKeys) get(ID string) (string, error) {
	keysAge := int(time.Now().UTC().Sub(k.refreshedAt).Seconds())
	if keysAge > k.maxAge {
		k.refresh()
	}

	return k.keys[ID], nil
}

func (k *pubKeys) refresh() error {
	log.Info("Refresh pub keys...")

	resp, err := k.client.Get(pubKeyURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&k.keys)
	if err != nil {
		return err
	}
	cacheControl := resp.Header["Cache-Control"]

	k.m.Lock()
	defer k.m.Unlock()
	for _, v := range cacheControl {
		match := cacheControlMaxAgeRe.FindStringSubmatch(v)
		if match != nil {
			maxAge, err := strconv.Atoi(match[1])
			if err != nil {
				return err
			}
			k.maxAge = maxAge
			break
		}
	}
	k.refreshedAt = time.Now().UTC()
	log.WithFields(log.Fields{
		"refreshedAt": k.refreshedAt,
		"maxAge":      k.maxAge,
	}).Info("Successfully refreshed pub keys")
	return nil
}
