package service

import (
	"context"
	"errors"
	"sync"

	"github.com/TranQuocToan1996/go-pcBookgRPC/pb"
)

var (
	ErrAlreadyExist = errors.New("already exist")
	ErrNotExist     = errors.New("not exist")
)

type LaptopStore interface {
	Save(context.Context, *pb.Laptop) error
	Find(context.Context, string) (*pb.Laptop, error)
}

type InMemoryLaptopStore struct {
	data  map[string]*pb.Laptop
	mutex sync.RWMutex
}

func NewInMemoryLaptopStore() *InMemoryLaptopStore {
	return &InMemoryLaptopStore{
		data: make(map[string]*pb.Laptop),
	}
}

func (i *InMemoryLaptopStore) Save(ctx context.Context, laptop *pb.Laptop) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	if _, exist := i.data[laptop.Id]; exist {
		return ErrAlreadyExist
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		i.data[laptop.Id] = laptop
	}

	return nil
}

func (i *InMemoryLaptopStore) Find(ctx context.Context, id string) (*pb.Laptop, error) {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	if _, exist := i.data[id]; !exist {
		return nil, ErrNotExist
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return i.data[id], nil
	}
}
