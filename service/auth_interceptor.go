package service

import (
	"context"
	"log"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthInterceptor struct {
	jwtManager      *JWTManager
	accessibleRoles map[string][]string
}

func NewAuthInterceptor(jwtManager *JWTManager,
	accessibleRoles map[string][]string) *AuthInterceptor {
	return &AuthInterceptor{
		jwtManager:      jwtManager,
		accessibleRoles: accessibleRoles,
	}
}

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {

		log.Println("Unary interceptor", info.FullMethod)

		err = i.authorize(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func (i *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(server interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler) error {

		log.Println("Stream interceptor", info.FullMethod)

		err := i.authorize(stream.Context(), info.FullMethod)
		if err != nil {
			return err
		}

		return handler(server, stream)
	}
}

func (i *AuthInterceptor) authorize(ctx context.Context, method string) error {

	accessible, ok := i.accessibleRoles[method]
	if ok {
		md, exist := metadata.FromIncomingContext(ctx)
		if !exist {
			return status.Errorf(codes.Unauthenticated, "not yet sent token")
		}

		tokens := md["authorization"]
		if tokens == nil {
			return status.Errorf(codes.Unauthenticated, "token empty")
		}

		token := tokens[0]
		claims, err := i.jwtManager.Verify(token)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		for _, role := range accessible {
			if strings.EqualFold(role, claims.Role) {
				return nil
			}
		}

		return status.Errorf(codes.PermissionDenied, "user dont have permission")
	}

	return nil
}
