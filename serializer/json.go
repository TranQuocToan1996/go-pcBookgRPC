package serializer

import (
	"fmt"
	"os"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func WriteProtoBufToJsonFile(mess proto.Message, filename string) error {
	data, err := ProtbufToStringByteJson(mess)
	if err != nil {
		return fmt.Errorf("cant write byte json: %w", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("cant write json data to file: %w", err)
	}
	return nil
}

func ProtbufToStringJson(mess proto.Message) string {
	marshaler := protojson.MarshalOptions{
		Indent:        "\t",
		UseProtoNames: false,
	}

	return marshaler.Format(mess)
}

func ProtbufToStringByteJson(mess proto.Message) ([]byte, error) {
	marshaler := protojson.MarshalOptions{
		Indent:          "\t",  // Format json tab
		Multiline:       true,  // format json
		EmitUnpopulated: false, // omitempty
		AllowPartial:    false, // strictly include require fields
		UseProtoNames:   false, // format key name in json.
		UseEnumNumbers:  false, // Use name enum or use the value (interger), false is use key
	}

	return marshaler.Marshal(mess)
}
