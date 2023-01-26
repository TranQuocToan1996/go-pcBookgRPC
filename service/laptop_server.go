package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"

	"github.com/TranQuocToan1996/go-pcBookgRPC/pb"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	MaxChunkSize = 1 << 20
)

type LaptopServer struct {
	pb.UnimplementedLaptopServiceServer

	laptopStore LaptopStore
	imageStore  ImageStore
	ratingStore RatingStore
}

func NewLaptopServer(laptopStore LaptopStore, imageStore ImageStore, ratingStore RatingStore) *LaptopServer {
	return &LaptopServer{
		laptopStore: laptopStore,
		imageStore:  imageStore,
		ratingStore: ratingStore,
	}
}

func (s *LaptopServer) CreateLaptop(
	ctx context.Context,
	req *pb.CreateLaptopRequest,
) (*pb.CreateLaptopResponse, error) {
	laptop := req.GetLaptop()
	log.Printf("receive createLaptop req with id: %v", laptop.Id)

	if len(laptop.Id) > 0 {
		_, err := uuid.Parse(laptop.Id)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "laptop ID is not uuid format: %v", err.Error())
		}
	} else {
		id, err := uuid.NewRandom()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "empty laptop ID in req and server fail to create one: %v", err.Error())
		}

		laptop.Id = id.String()
	}

	err := contextError(ctx)
	if err != nil {
		return nil, err
	}

	err = s.laptopStore.Save(ctx, laptop)
	if err != nil {
		code := codes.Internal
		if errors.Is(err, ErrAlreadyExist) {
			code = codes.AlreadyExists
		}

		return nil, status.Errorf(code, "cant save laptop obj: %v", err.Error())
	}

	log.Printf("saved laptop id: %v", laptop.Id)

	return &pb.CreateLaptopResponse{
		Id: laptop.Id,
	}, nil
}

func (s *LaptopServer) SearchLaptop(req *pb.SearchLaptopRequest,
	stream pb.LaptopService_SearchLaptopServer) error {
	filter := req.GetFilter()
	log.Printf("receiving")

	err := s.laptopStore.Search(stream.Context(),
		filter,
		func(laptop *pb.Laptop) error {
			res := &pb.SearchLaptopResponse{Laptop: laptop}

			err := stream.Send(res)
			if err != nil {
				return err
			}

			log.Printf("send laptop with id %v", laptop.GetId())

			return nil
		})
	if err != nil {
		return status.Errorf(codes.Internal, "unexpect error %v", err)
	}

	return nil
}

func (s *LaptopServer) UploadImage(stream pb.LaptopService_UploadImageServer) error {

	req, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.Unknown, "cant receive image %s", err.Error())
	}

	laptopID := req.GetInfo().GetLaptopId()
	imageType := req.GetInfo().GetImageType()
	log.Printf("receiver an image upload request from laptopID %s with type %s",
		laptopID, imageType)

	laptop, err := s.laptopStore.Find(context.TODO(), laptopID)
	if err != nil {
		return status.Errorf(codes.Internal, "error when finding laptop %s", err.Error())
	}
	if laptop == nil {
		return status.Errorf(codes.NotFound, "not found laptop %s", laptopID)
	}

	imageData := bytes.NewBuffer(nil)
	imageSize := 0

	for log.Print("Start saving"); ; {
		log.Print("receiving data chunk at chunk size ", imageSize)

		err := contextError(stream.Context())
		if err != nil {
			return err
		}

		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("no more image data")
			break
		}
		if err != nil {
			return status.Errorf(codes.Unknown, "cant receive data %v", err)
		}

		chunk := req.GetChunkData()
		size := len(chunk)
		if size > MaxChunkSize {
			return status.Errorf(codes.InvalidArgument, "too big image chunk")
		}
		imageSize += size

		_, err = imageData.Write(chunk)
		if err != nil {
			return status.Errorf(codes.Internal, "error when write data %v", err)
		}
	}

	imageID, err := s.imageStore.Save(laptopID, imageType, *imageData)
	if err != nil {
		return status.Errorf(codes.Internal, "error when save file %v", err)
	}

	res := &pb.UploadImageResponse{
		Id:   imageID,
		Size: uint32(imageSize),
	}

	err = stream.SendAndClose(res)
	if err != nil {
		return status.Errorf(codes.Internal, "error when sending response %v", err)
	}

	return nil
}

func (s *LaptopServer) RateLaptop(stream pb.LaptopService_RateLaptopServer) error {
	for {
		err := contextError(stream.Context())
		if err != nil {
			return err
		}

		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("rate laptop req has no more data")
			break
		}
		if err != nil {
			return status.Errorf(codes.Unknown, "cant receive rating request: %v", err)
		}

		laptopID, score := req.LaptopId, req.Score
		log.Printf("[Rating laptop] got laptop ID %v and score %v", laptopID, score)
		found, err := s.laptopStore.Find(stream.Context(), laptopID)
		if err != nil {
			return status.Errorf(codes.Internal, "[Rating lapttop] error when find laptopID %v: %v", laptopID, err)
		}
		if found == nil {
			return status.Errorf(codes.NotFound, "[Rating lapttop] cant find laptop with id %v", laptopID)
		}

		rating, err := s.ratingStore.Add(laptopID, score)
		if err != nil {
			return status.Errorf(codes.Internal, "[Rating lapttop] cant add rate to store %v: %v", laptopID, err)
		}

		res := &pb.RateLaptopResponse{
			LaptopId:     laptopID,
			RatedCount:   rating.Count,
			AverageScore: rating.Sum / float64(rating.Count),
		}

		err = stream.Send(res)
		if err != nil {
			return status.Errorf(codes.Internal, "[Rating lapttop] unexpect error when sending back resp %v: %v", laptopID, err)
		}
	}

	return nil
}

func contextError(ctx context.Context) error {
	if ctx.Err() == context.DeadlineExceeded {
		log.Print("deadline exceed")
		return status.Error(codes.DeadlineExceeded, "deadline exceed with laptop")
	}

	if ctx.Err() == context.Canceled {
		log.Print("request cancel by client")
		return status.Error(codes.Canceled, "request cancel by client")
	}

	return nil
}
