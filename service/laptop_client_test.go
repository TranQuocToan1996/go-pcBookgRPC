package service_test

import (
	"context"
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

	server, address, err := startTestLaptopServer()
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

func startTestLaptopServer() (server *service.LaptopServer, address string, err error) {
	server = service.NewLaptopServer(service.NewInMemoryLaptopStore())
	gprcServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(gprcServer, server)

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
