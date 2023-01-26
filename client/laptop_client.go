package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/TranQuocToan1996/go-pcBookgRPC/pb"
	"github.com/TranQuocToan1996/go-pcBookgRPC/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LaptopRate struct {
	LaptopID string
	Score    float64
}

type LaptopClient struct {
	service pb.LaptopServiceClient
}

func NewLaptopClient(service pb.LaptopServiceClient) *LaptopClient {
	return &LaptopClient{service}
}

func (c *LaptopClient) RateLaptop(list []LaptopRate) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	stream, err := c.service.RateLaptop(ctx)
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
			LaptopId: obj.LaptopID,
			Score:    obj.Score,
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

func (c *LaptopClient) UploadImage(laptopID string, path string) *pb.UploadImageResponse {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	stream, err := c.service.UploadImage(ctx)
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

func (c *LaptopClient) SearchLaptop(filter *pb.Filter) {
	log.Print("search filter", filter)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	req := &pb.SearchLaptopRequest{Filter: filter}
	stream, err := c.service.SearchLaptop(ctx, req)
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

func (c *LaptopClient) CreateLaptop(laptop *pb.Laptop) {

	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	res, err := c.service.CreateLaptop(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			log.Print("laptop already exist")
		} else {
			log.Fatal("cant create laptop ", err)
		}
	}

	log.Printf("create laptop with id %v", res.Id)
}
