package main

import (
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
	imageFolder       = "./img"
	secretKey         = "Need to generate key"
	tokenDuration     = time.Hour
	adminRole         = "admin"
	userRole          = "user"
	laptopServicePath = "/pb.LaptopService/"
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

		ClientAuth: tls.RequireAndVerifyClientCert, // mutual TLS
		ClientCAs:  certPool,                       // mutual TLS

		// ClientAuth:   tls.NoClientCert, // Client no need to auth
	}

	return credentials.NewTLS(config), nil
}

func main() {
	port := flag.String("serverport", "8080", "server port")
	enableTLS := flag.Bool("tls", false, "enable SSL/TLS")
	flag.Parse()
	log.Printf("starting server on port: %v, TLS: %v", *port, *enableTLS)

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

	serverOTPs := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.Unary()),
		grpc.StreamInterceptor(interceptor.Stream()),
	}

	if *enableTLS {
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

	address := fmt.Sprintf(":%v", *port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(err)
	}

	log.Print(grpcServer.Serve(lis))

}
