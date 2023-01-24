package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/TranQuocToan1996/go-pcBookgRPC/pb"
	"github.com/TranQuocToan1996/go-pcBookgRPC/service"
	"google.golang.org/grpc"
)

const (
	imageFolder = "./img"
)

func main() {
	port := flag.String("serverport", "8080", "server port")
	flag.Parse()
	log.Printf("starting server on port %v", *port)

	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore(imageFolder)

	server := service.NewLaptopServer(laptopStore, imageStore)
	grpcServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(grpcServer, server)

	address := fmt.Sprintf(":%v", *port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(err)
	}

	log.Print(grpcServer.Serve(lis))

}
