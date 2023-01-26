package service_test

import (
	"context"
	"testing"

	"github.com/TranQuocToan1996/go-pcBookgRPC/pb"
	"github.com/TranQuocToan1996/go-pcBookgRPC/sample"
	"github.com/TranQuocToan1996/go-pcBookgRPC/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServerSaveLaptop(t *testing.T) {
	t.Parallel()

	laptopNoID := sample.NewLaptop()
	laptopNoID.Id = ""

	laptopInvalidID := sample.NewLaptop()
	laptopInvalidID.Id = "invalid UUID"

	laptopDuplicateID := sample.NewLaptop()
	duplicateID := uuid.New().String()
	laptopDuplicateID.Id = duplicateID
	storeDuplicateID := service.NewInMemoryLaptopStore()
	err := storeDuplicateID.Save(context.TODO(), laptopDuplicateID)
	require.Nil(t, err)

	tests := []struct {
		name   string
		laptop *pb.Laptop
		store  service.LaptopStore
		code   codes.Code
	}{
		{
			name:   "success_reqWithID",
			laptop: sample.NewLaptop(),
			store:  service.NewInMemoryLaptopStore(),
			code:   codes.OK,
		},
		{
			name:   "success_reqNoID",
			laptop: laptopNoID,
			store:  service.NewInMemoryLaptopStore(),
			code:   codes.OK,
		},
		{
			name:   "success_reqInvalidID",
			laptop: laptopInvalidID,
			store:  service.NewInMemoryLaptopStore(),
			code:   codes.InvalidArgument,
		},
		{
			name:   "success_idAlreadyExist",
			laptop: laptopInvalidID,
			store:  service.NewInMemoryLaptopStore(),
			code:   codes.InvalidArgument,
		},
	}

	for _, test := range tests {
		// Prevent tests goroutines access same to the last test
		// By assgin to block var
		tc := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			req := &pb.CreateLaptopRequest{Laptop: tc.laptop}
			server := service.NewLaptopServer(tc.store, nil, nil)
			res, err := server.CreateLaptop(context.TODO(), req)
			if tc.code == codes.OK {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.Equal(t, tc.laptop.Id, res.Id)
			} else {
				require.Error(t, err)
				require.Nil(t, res)
				code, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tc.code, code.Code())
			}
		})
	}
}
