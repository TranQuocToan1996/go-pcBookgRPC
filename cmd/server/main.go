package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/TranQuocToan1996/go-pcBookgRPC/pb"
	"github.com/TranQuocToan1996/go-pcBookgRPC/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	imageFolder = "./img"
)

const (
	admin = "admin"
	user  = "user"
)

func unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

	log.Println("Unary interceptor", info.FullMethod)

	return handler(ctx, req)
}

func streamInterceptor(server interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	log.Println("Stream interceptor", info.FullMethod)

	return handler(server, stream)
}

func seedUsers(userStore service.UserStore) error {
	err := createUsers(userStore, "admin1", "admin1", admin)
	if err != nil {
		return err
	}
	return createUsers(userStore, "user1", "user1", user)
}

func createUsers(userStore service.UserStore,
	username, rawPw, role string) error {

	user, err := service.NewUser(username, rawPw, role)
	if err != nil {
		return err
	}
	return userStore.Save(user)
}

func accessibleRoles() map[string][]string {

	const (
		laptopServicePath = "/pb.LaptopService/"
	)

	return map[string][]string{
		laptopServicePath + "CreateLaptop": {admin},
		laptopServicePath + "UploadImage":  {admin},
		laptopServicePath + "RateLaptop":   {admin, user},
	}
}

const (
	secretKey     = "Need to generate key"
	tokenDuration = time.Hour
)

func main() {
	port := flag.String("serverport", "8080", "server port")
	flag.Parse()
	log.Printf("starting server on port %v", *port)

	userStore := service.NewInMemoryUserStore()
	err := seedUsers(userStore)
	if err != nil {
		log.Fatal(err)
	}
	jwtManager := service.NewJWTManager(secretKey, tokenDuration)
	interceptor := service.NewAuthInterceptor(jwtManager, accessibleRoles())
	authServer := service.NewAuthServer(userStore, jwtManager)

	laptopStore := service.NewInMemoryLaptopStore()
	ratingStore := service.NewInMemoryRatingStore()
	imageStore := service.NewDiskImageStore(imageFolder)

	laptopServer := service.NewLaptopServer(laptopStore, imageStore, ratingStore)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.Unary()),
		grpc.StreamInterceptor(interceptor.Stream()),
	)

	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)
	pb.RegisterAuthServiceServer(grpcServer, authServer)
	reflection.Register(grpcServer)

	address := fmt.Sprintf(":%v", *port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(err)
	}

	log.Print(grpcServer.Serve(lis))

}
