package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/TranQuocToan1996/go-pcBookgRPC/pb"
	"github.com/TranQuocToan1996/go-pcBookgRPC/service"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

const (
	imageFolder       = "./img"
	secretKey         = "Need to generate key"
	tokenDuration     = time.Hour
	adminRole         = "admin"
	userRole          = "user"
	laptopServicePath = "/pb.LaptopService/"
	serverCert        = "cert/server-cert.pem"
	serverKey         = "cert/server-key.pem"
	caCert            = "cert/ca-cert.pem"
)

func seedUsers(userStore service.UserStore) error {
	err := createUsers(userStore, "admin1", "admin1", adminRole)
	if err != nil {
		return err
	}
	return createUsers(userStore, "user1", "user1", userRole)
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
	return map[string][]string{
		laptopServicePath + "CreateLaptop": {adminRole},
		laptopServicePath + "UploadImage":  {adminRole},
		laptopServicePath + "RateLaptop":   {adminRole, userRole},
	}
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	serverCert, err := tls.LoadX509KeyPair(serverCert,
		secretKey)
	if err != nil {
		return nil, err
	}

	cert, err := os.ReadFile(caCert)
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

		ClientAuth: tls.RequireAndVerifyClientCert, // mutual TLS
		ClientCAs:  certPool,                       // mutual TLS

		// ClientAuth:   tls.NoClientCert, // Client no need to auth
	}

	return credentials.NewTLS(config), nil
}

func runGRPCServer(authServer pb.AuthServiceServer,
	laptopServer pb.LaptopServiceServer,
	jwtManager *service.JWTManager,
	enableTLS bool, listener net.Listener) error {

	interceptor := service.NewAuthInterceptor(jwtManager, accessibleRoles())

	serverOTPs := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.Unary()),
		grpc.StreamInterceptor(interceptor.Stream()),
	}

	if enableTLS {
		loadTLSCredentials, err := loadTLSCredentials()
		if err != nil {
			log.Fatal(err)
		}
		serverOTPs = append(serverOTPs, grpc.Creds(loadTLSCredentials))
	}

	grpcServer := grpc.NewServer(serverOTPs...)

	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)
	pb.RegisterAuthServiceServer(grpcServer, authServer)
	reflection.Register(grpcServer)

	log.Println("starting GRPC server")

	return grpcServer.Serve(listener)
}

func runRESTServer(authServer pb.AuthServiceServer,
	laptopServer pb.LaptopServiceServer,
	jwtManager *service.JWTManager,
	enableTLS bool, listener net.Listener) error {

	mux := runtime.NewServeMux()
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*15)
	defer cancel()

	err := pb.RegisterAuthServiceHandlerServer(ctx, mux, authServer)
	if err != nil {
		return err
	}

	err = pb.RegisterLaptopServiceHandlerServer(ctx, mux, laptopServer)
	if err != nil {
		return err
	}

	log.Println("starting REST server")
	if enableTLS {
		return http.ServeTLS(listener, mux, serverCert, serverKey)
	}

	return http.Serve(listener, mux)

}

func main() {
	port := flag.String("serverport", "8080", "server port")
	enableTLS := flag.Bool("tls", false, "enable SSL/TLS")
	restServer := flag.Bool("rest", false, "enable REST instead of GRPC")
	flag.Parse()
	log.Printf("starting server on port: %v, TLS: %v", *port, *enableTLS)

	userStore := service.NewInMemoryUserStore()
	err := seedUsers(userStore)
	if err != nil {
		log.Fatal(err)
	}
	jwtManager := service.NewJWTManager(secretKey, tokenDuration)

	address := fmt.Sprintf(":%v", *port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(err)
	}

	laptopStore := service.NewInMemoryLaptopStore()
	ratingStore := service.NewInMemoryRatingStore()
	imageStore := service.NewDiskImageStore(imageFolder)

	authServer := service.NewAuthServer(userStore, jwtManager)
	laptopServer := service.NewLaptopServer(laptopStore, imageStore, ratingStore)

	if *restServer {
		err = runRESTServer(authServer, laptopServer, jwtManager, *enableTLS, lis)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err = runGRPCServer(authServer, laptopServer, jwtManager, *enableTLS, lis)
		if err != nil {
			log.Fatal(err)
		}
	}

}
