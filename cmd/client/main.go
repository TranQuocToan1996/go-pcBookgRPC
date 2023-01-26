package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/TranQuocToan1996/go-pcBookgRPC/pb"
	"github.com/TranQuocToan1996/go-pcBookgRPC/sample"
	"github.com/TranQuocToan1996/go-pcBookgRPC/service"
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
	conn, err := grpc.Dial(adddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	laptopClient := pb.NewLaptopServiceClient(conn)
	testRateLaptop(laptopClient)

}

type laptopRate struct {
	laptopID string
	score    float64
}

func rateLaptop(laptopClient pb.LaptopServiceClient, list []laptopRate) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	stream, err := laptopClient.RateLaptop(ctx)
	if err != nil {
		return err
	}

	chanErr := make(chan error)
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				log.Print("[RateLaptop] EOF no more response from client.")
				chanErr <- nil
				return
			}
			if err != nil {
				chanErr <- err
				return
			}
			log.Print("receive resp ", res)
		}
	}()

	for _, obj := range list {
		req := &pb.RateLaptopRequest{
			LaptopId: obj.laptopID,
			Score:    obj.score,
		}

		err := stream.Send(req)
		if err != nil {
			return fmt.Errorf("cant send stream req: %v - %v", err, stream.RecvMsg(nil))
		}

		log.Print("send req", req)
	}

	err = stream.CloseSend()
	if err != nil {
		return fmt.Errorf("cant send close request: %v", err)
	}

	return <-chanErr

}

func testRateLaptop(laptopClient pb.LaptopServiceClient) {
	list := [3]laptopRate{}
	for i := 0; i < len(list); i++ {
		laptop := sample.NewLaptop()
		list[i].laptopID = laptop.GetId()
		list[i].score = sample.RandomLaptopScore()
		createLaptop(laptopClient, laptop)
	}

	for i := 0; i < len(list); i++ {
		err := rateLaptop(laptopClient, list[:])
		if err != nil {
			log.Fatal(err)
		}
	}

}

func uploadImage(laptopClient pb.LaptopServiceClient,
	laptopID string, path string) *pb.UploadImageResponse {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	stream, err := laptopClient.UploadImage(ctx)
	if err != nil {
		log.Fatal(err)
	}
	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{
			Info: &pb.ImageInfo{
				LaptopId:  laptopID,
				ImageType: filepath.Ext(path),
			},
		},
	}

	err = stream.Send(req)
	if err != nil {
		err2 := stream.RecvMsg(nil)
		log.Fatal(err, err2)
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, service.MaxChunkSize)

	size := 0

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil
		}

		size += n

		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}
		err = stream.Send(req)
		if err != nil {
			err2 := stream.RecvMsg(nil)
			log.Fatal(err, err2)
		}
		log.Println("Readsize:", size)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("receive id %v and size %v from server reponse", res.Id, res.Size)

	return res
}

func testCreateLaptop(laptopClient pb.LaptopServiceClient) {
	createLaptop(laptopClient, sample.NewLaptop())
}

func testUploadImage(laptopClient pb.LaptopServiceClient) {
	laptop := sample.NewLaptop()
	createLaptop(laptopClient, laptop)
	uploadImage(laptopClient, laptop.GetId(), "tmp/laptop.png")
}

func testSearchLaptop(laptopClient pb.LaptopServiceClient) {
	for i := 0; i < 10; i++ {
		createLaptop(laptopClient, sample.NewLaptop())
	}

	filter := &pb.Filter{
		MaxPriceUsd: 3000,
		MinCpuCores: 4,
		MinCpuGhz:   2.5,
		MinRam:      &pb.Memory{Value: 8, Unit: pb.Memory_GIGABYTE},
	}

	searchLaptop(laptopClient, filter)
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

func createLaptop(laptopClient pb.LaptopServiceClient, laptop *pb.Laptop) {

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
