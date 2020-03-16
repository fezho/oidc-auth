package dex

import (
	"context"
	"fmt"
	"github.com/fezho/oidc-auth-service/dex/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Dexy is the proxy for calling dex api with gRPC
type Dexy struct {
	client api.DexClient
	conn   *grpc.ClientConn
}

func New(target string) (*Dexy, error) {
	conn, err := grpc.Dial(target, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &Dexy{
		client: api.NewDexClient(conn),
		conn:   conn,
	}, nil
}

func NewWithCredentials(target, caPath string) (*Dexy, error) {
	creds, err := credentials.NewClientTLSFromFile(caPath, "")
	if err != nil {
		return nil, fmt.Errorf("load dex cert: %v", err)
	}

	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("dial: %v", err)
	}

	return &Dexy{
		client: api.NewDexClient(conn),
		conn:   conn,
	}, nil
}

func (d *Dexy) CreateClient(id, name string, redirectURI string) (string, error) {
	// TODO: generate secret
	secret := "test-auth-app-secret"
	req := &api.CreateClientReq{
		Client: &api.Client{
			Id:           id,
			Secret:       secret,
			RedirectUris: []string{redirectURI},
			Name:         name,
		},
	}
	res, err := d.client.CreateClient(context.TODO(), req)
	if err != nil {
		return "", err
	}
	if res.AlreadyExists {
		//return "", fmt.Errorf("client %q already exists.", id)
	}
	return secret, nil
}

// TODO: add others

func (d *Dexy) Close() error {
	return d.conn.Close()
}
