package client

import (
	"context"
	"time"

	"github.com/TranQuocToan1996/go-pcBookgRPC/pb"
	"google.golang.org/grpc"
)

type AuthClient struct {
	service            pb.AuthServiceClient
	username, password string
}

func (a *AuthClient) Login() (string, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	req := &pb.LoginRequest{
		Username: a.username,
		Password: a.password,
	}

	res, err := a.service.Login(ctx, req)
	if err != nil {
		return "", err
	}

	return res.AccessToken, nil
}

func NewAuthClient(cc *grpc.ClientConn,
	username, password string) *AuthClient {
	service := pb.NewAuthServiceClient(cc)
	return &AuthClient{
		service:  service,
		username: username,
		password: password,
	}
}
