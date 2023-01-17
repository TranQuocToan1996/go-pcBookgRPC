package serializer

import (
	"fmt"
	"os"
	"reflect"

	"google.golang.org/protobuf/proto"
)

func WriteProtoBufToByte(mess proto.Message, filename string) error {
	data, err := proto.Marshal(mess)
	if err != nil {
		return fmt.Errorf("can't marshal proto message: %w", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("can't writefile: %w", err)
	}

	return nil
}

func ReadProtoBufBin(mess proto.Message, filename string) error {
	if reflect.ValueOf(mess).Kind() != reflect.Pointer {
		return fmt.Errorf("the mess var must be passing in type of pointer to get data from unmarshal")
	}

	buf, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("can't read file: %w", err)
	}

	err = proto.Unmarshal(buf, mess)
	if err != nil {
		return fmt.Errorf("can't Unmarshal %w", err)
	}

	return nil
}


