package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/TranQuocToan1996/go-pcBookgRPC/pb"
	"github.com/TranQuocToan1996/go-pcBookgRPC/sample"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
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
	for i := 0; i < 10; i++ {
		createLaptop(laptopClient)
	}

	filter := &pb.Filter{
		MaxPriceUsd: 3000,
		MinCpuCores: 4,
		MinCpuGhz:   2.5,
		MinRam:      &pb.Memory{Value: 8, Unit: pb.Memory_GIGABYTE},
	}

	searchLaptop(laptopClient, filter)

	// ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	// defer cancel()
	// res, err := laptopClient.CreateLaptop(ctx, &pb.CreateLaptopRequest{Laptop: sample.NewLaptop()})
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// log.Printf("create laptop with id: %v", res.Id)
}

func searchLaptop(laptopClient pb.LaptopServiceClient, filter *pb.Filter) {
	log.Print("search filter", filter)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	req := &pb.SearchLaptopRequest{Filter: filter}
	stream, err := laptopClient.SearchLaptop(ctx, req)
	if err != nil {
		log.Fatal("cant search laptop", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatal("cant receive stream resp", err)
		}
		laptop := res.GetLaptop()
		log.Printf("found laptop %v - %v", laptop.Id, laptop.Brand)
	}
}

func createLaptop(laptopClient pb.LaptopServiceClient) {
	laptop := sample.NewLaptop()

	laptop.Id = ""
	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	res, err := laptopClient.CreateLaptop(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			log.Print("laptop already exist")
		} else {
			log.Fatal("cant create laptop", err)
		}
	}

	log.Printf("create laptop with id %v", res.Id)
}
