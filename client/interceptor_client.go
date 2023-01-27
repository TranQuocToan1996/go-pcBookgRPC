package client

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type AuthInterceptor struct {
	client      *AuthClient
	authMethods map[string]bool
	accessToken string
}

func NewAuthInterceptor(
	client *AuthClient,
	authMethods map[string]bool,
	refreshTokenDuration time.Duration,
) (*AuthInterceptor, error) {
	interceptor := &AuthInterceptor{
		client:      client,
		authMethods: authMethods,
	}

	err := interceptor.scheduleRefreshToken(refreshTokenDuration)
	if err != nil {
		return nil, err
	}

	return interceptor, nil
}

func (i *AuthInterceptor) scheduleRefreshToken(refreshTokenDuration time.Duration) error {
	err := i.refreshToken()
	if err != nil {
		return err
	}

	go func() {
		ticket := time.NewTicker(refreshTokenDuration)
		refreshNow := make(chan time.Time)
		retry := 0
		for {
			if retry > 10 {
				log.Fatal("exceed 10 time retry refresh token")
			}

			select {
			case <-ticket.C:
				err := i.refreshToken()
				if err != nil {
					log.Println(err)
					refreshNow <- time.Now()
					retry++
				} else {
					retry = 0
				}
			case <-refreshNow:
				err := i.refreshToken()
				if err != nil {
					time.Sleep(time.Second * 5)
					log.Println(err)
					refreshNow <- time.Now()
					retry++
				} else {
					retry = 0
				}
			}
		}
	}()

	return nil
}

func (i *AuthInterceptor) refreshToken() error {
	token, err := i.client.Login()
	if err != nil {
		return err
	}

	i.accessToken = token
	log.Printf("token refresh: %v", token)
	return nil
}

func (i *AuthInterceptor) Unary() grpc.UnaryClientInterceptor {
	return func(ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {

		log.Println("Unary client" + method)
		if i.authMethods[method] {
			return invoker(i.attachToken(ctx), method, req, reply, cc, opts...)
		}

		return invoker(i.attachToken(ctx), method, req, reply, cc, opts...)
	}
}

func (i *AuthInterceptor) attachToken(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", i.accessToken)
}

func (i *AuthInterceptor) Stream() grpc.StreamClientInterceptor {
	return func(ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption) (grpc.ClientStream, error) {

		log.Println("Stream client" + method)

		if i.authMethods[method] {
			return streamer(i.attachToken(ctx), desc, cc, method, opts...)
		}

		return streamer(i.attachToken(ctx), desc, cc, method, opts...)
	}
}
w