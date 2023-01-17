package service

import (
	"context"
	"errors"
	"log"

	"github.com/TranQuocToan1996/go-pcBookgRPC/pb"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LaptopServer struct {
	pb.UnimplementedLaptopServiceServer

	Store LaptopStore
}

func NewLaptopServer(store LaptopStore) *LaptopServer {
	return &LaptopServer{
		Store: store,
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

	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("deadline exceed with laptop id %v", laptop.Id)
		return nil, status.Error(codes.DeadlineExceeded, "deadline exceed with laptop")
	}

	if ctx.Err() == context.Canceled {
		log.Printf("request cancel by client with laptop id %v", laptop.Id)
		return nil, status.Error(codes.Canceled, "request cancel by client")
	}

	err := s.Store.Save(ctx, laptop)
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
