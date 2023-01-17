package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/TranQuocToan1996/go-pcBookgRPC/pb"
	"github.com/TranQuocToan1996/go-pcBookgRPC/sample"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	port := flag.String("serverport", "8080", "server port")
	flag.Parse()
	log.Printf("calling on port %v", *port)
	adddress := fmt.Sprintf("0.0.0.0:%v", *port)
	conn, err := grpc.Dial(adddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	laptopClient := pb.NewLaptopServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	res, err := laptopClient.CreateLaptop(ctx, &pb.CreateLaptopRequest{Laptop: sample.NewLaptop()})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("create laptop with id: %v", res.Id)
}
