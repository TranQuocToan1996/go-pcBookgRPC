package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/TranQuocToan1996/go-pcBookgRPC/pb"
	"github.com/TranQuocToan1996/go-pcBookgRPC/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	serverCert, err := tls.LoadX509KeyPair("cert/server-cert.pem",
		"cert/server-key.pem")
	if err != nil {
		return nil, err
	}

	cert, err := os.ReadFile("cert/ca-cert.pem")
	if err != nil {
		return nil, err
	}

	// same CA for client and server
	certPool := x509.NewCertPool()

	if !certPool.AppendCertsFromPEM(cert) {
		return nil, fmt.Errorf("fail server cert")
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		// ClientAuth:   tls.NoClientCert, // When server not care about client call
		ClientAuth: tls.RequireAndVerifyClientCert, // mutual TLS
		ClientCAs:  certPool,                       // mutual TLS
	}

	return credentials.NewTLS(config), nil
}

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

	loadTLSCredentials, err := loadTLSCredentials()
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer(
		grpc.Creds(loadTLSCredentials),
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
