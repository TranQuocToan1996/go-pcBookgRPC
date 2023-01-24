package service_test

import (
	"context"
	"io"
	"log"
	"net"
	"testing"

	"github.com/TranQuocToan1996/go-pcBookgRPC/pb"
	"github.com/TranQuocToan1996/go-pcBookgRPC/sample"
	"github.com/TranQuocToan1996/go-pcBookgRPC/service"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

func TestClientCreateLaptop(t *testing.T) {
	t.Parallel()

	server, address, err := startTestLaptopServer(service.NewInMemoryLaptopStore())
	require.NoError(t, err)
	client, err := newClientLaptop(address)
	require.NoError(t, err)

	laptop := sample.NewLaptop()
	expectedID := laptop.Id
	req := &pb.CreateLaptopRequest{Laptop: laptop}

	res, err := client.CreateLaptop(context.TODO(), req)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, expectedID, res.Id)

	laptopFromBD, err := server.Store.Find(context.TODO(), res.Id)
	require.NoError(t, err)
	require.NotNil(t, laptopFromBD)

	// Must use proto.Equal because inside pb.Laptop struct has internal fields (IE sizeCache)
	require.True(t, proto.Equal(laptopFromBD, laptop))

}

func startTestLaptopServer(store service.LaptopStore) (server *service.LaptopServer, address string, err error) {
	server = service.NewLaptopServer(service.NewInMemoryLaptopStore())
	gprcServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(gprcServer, server)

	server.Store = store

	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, "", err
	}

	go func() {
		log.Print(gprcServer.Serve(lis))
	}()
	return server, lis.Addr().String(), nil
}

func newClientLaptop(address string) (pb.LaptopServiceClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return pb.NewLaptopServiceClient(conn), nil
}

func TestClientSearchLaptop(t *testing.T) {
	t.Parallel()

	const noOfLaptops = 6
	var (
		expectedIDs = make(map[string]bool)
		store       = service.NewInMemoryLaptopStore()
	)

	for i := 0; i < noOfLaptops; i++ {
		laptop := sample.NewLaptop()
		err := store.Save(context.Background(), laptop)
		require.NoError(t, err)

		switch i {
		case 0:
			laptop.PriceUsd = 2500
		case 1:
			laptop.Cpu.NumberCores = 2
		case 2:
			laptop.Cpu.MinGhz = 2
		case 3:
			laptop.Ram = &pb.Memory{Value: 1024, Unit: pb.Memory_MEGABYTE}
		case 4:
			laptop.PriceUsd = 1000
			laptop.Cpu.NumberCores = 4
			laptop.Cpu.MinGhz = 2.2
			laptop.Ram = &pb.Memory{Value: 8, Unit: pb.Memory_GIGABYTE}
			expectedIDs[laptop.Id] = true
		case 5:
			laptop.PriceUsd = 2000
			laptop.Cpu.NumberCores = 4
			laptop.Cpu.MinGhz = 2.2
			laptop.Ram = &pb.Memory{Value: 8, Unit: pb.Memory_GIGABYTE}
			expectedIDs[laptop.Id] = true
		}
	}

	_, address, err := startTestLaptopServer(store)
	require.NoError(t, err)

	client, err := newClientLaptop(address)
	require.NoError(t, err)

	filter := &pb.Filter{
		MaxPriceUsd: 2000,
		MinCpuCores: 4,
		MinCpuGhz:   2.2,
		MinRam:      &pb.Memory{Value: 8, Unit: pb.Memory_GIGABYTE},
	}

	req := &pb.SearchLaptopRequest{Filter: filter}
	stream, err := client.SearchLaptop(context.Background(), req)
	require.NoError(t, err)

	var found int
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		require.Contains(t, expectedIDs, res.GetLaptop().GetId())
		found++
	}

	require.Equal(t, found, len(expectedIDs))

}
