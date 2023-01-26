package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/TranQuocToan1996/go-pcBookgRPC/client"
	"github.com/TranQuocToan1996/go-pcBookgRPC/pb"
	"github.com/TranQuocToan1996/go-pcBookgRPC/sample"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	username        = "admin1"
	password        = "admin1"
	refreshDuration = 30 * time.Second
)

func authMethods() map[string]bool {
	const (
		laptopServicePath = "/pb.LaptopService/"
	)

	return map[string]bool{
		laptopServicePath + "CreateLaptop": true,
		laptopServicePath + "UploadImage":  true,
		laptopServicePath + "RateLaptop":   true,
	}
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	clientCert, err := tls.LoadX509KeyPair("cert/client-cert.pem",
		"cert/client-key.pem")
	if err != nil {
		return nil, err
	}

	cert, err := os.ReadFile("cert/ca-cert.pem")
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()

	if !certPool.AppendCertsFromPEM(cert) {
		return nil, fmt.Errorf("fail server cert")
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}

	return credentials.NewTLS(config), nil
}

func main() {
	port := flag.String("serverport", "8080", "server port")
	flag.Parse()
	log.Printf("calling on port %v", *port)
	adddress := fmt.Sprintf("0.0.0.0:%v", *port)
	loadTLSCredentials, err := loadTLSCredentials()
	if err != nil {
		log.Fatal(err)
	}
	cc1, err := grpc.Dial(adddress,
		grpc.WithTransportCredentials(loadTLSCredentials))
	if err != nil {
		log.Fatal(err)
	}

	authClient := client.NewAuthClient(cc1, username, password)
	interceptor, err := client.NewAuthInterceptor(authClient, authMethods(), refreshDuration)
	if err != nil {
		log.Fatal(err)
	}

	cc2, err := grpc.Dial(adddress,
		grpc.WithTransportCredentials(loadTLSCredentials),
		grpc.WithUnaryInterceptor(interceptor.Unary()),
		grpc.WithStreamInterceptor(interceptor.Stream()),
	)
	if err != nil {
		log.Fatal(err)
	}

	laptopClient := client.NewLaptopClient(pb.NewLaptopServiceClient(cc2))
	testRateLaptop(laptopClient)

}

func testRateLaptop(laptopClient *client.LaptopClient) {
	list := [3]client.LaptopRate{}
	for i := 0; i < len(list); i++ {
		laptop := sample.NewLaptop()
		list[i].LaptopID = laptop.GetId()
		list[i].Score = sample.RandomLaptopScore()
		laptopClient.CreateLaptop(laptop)
	}

	for i := 0; i < len(list); i++ {
		err := laptopClient.RateLaptop(list[:])
		if err != nil {
			log.Fatal(err)
		}
	}

}

func testCreateLaptop(laptopClient *client.LaptopClient) {
	laptopClient.CreateLaptop(sample.NewLaptop())
}

func testUploadImage(laptopClient client.LaptopClient) {
	laptop := sample.NewLaptop()
	laptopClient.CreateLaptop(laptop)
	laptopClient.UploadImage(laptop.GetId(), "tmp/laptop.png")
}

func testSearchLaptop(laptopClient *client.LaptopClient) {
	for i := 0; i < 10; i++ {
		laptopClient.CreateLaptop(sample.NewLaptop())
	}

	filter := &pb.Filter{
		MaxPriceUsd: 3000,
		MinCpuCores: 4,
		MinCpuGhz:   2.5,
		MinRam:      &pb.Memory{Value: 8, Unit: pb.Memory_GIGABYTE},
	}

	laptopClient.SearchLaptop(filter)
}
