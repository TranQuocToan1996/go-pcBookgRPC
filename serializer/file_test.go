package serializer

import (
	"testing"

	"github.com/TranQuocToan1996/go-pcBookgRPC/pb"
	"github.com/TranQuocToan1996/go-pcBookgRPC/sample"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestFileMarshal(t *testing.T) {
	t.Parallel()

	binFile := "laptop.bin"
	jsonFile := "laptop.json"

	laptop1 := sample.NewLaptop()
	err := WriteProtoBufToByte(laptop1, binFile)
	require.NoError(t, err)

	laptop2 := &pb.Laptop{}
	err = ReadProtoBufBin(laptop2, binFile)
	require.NoError(t, err)
	require.True(t, proto.Equal(laptop1, laptop2))

	err = WriteProtoBufToJsonFile(laptop1, jsonFile)
	require.NoError(t, err)

}
