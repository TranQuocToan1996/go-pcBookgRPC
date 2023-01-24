package service

import (
	"context"
	"errors"
	"log"
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
	Search(context.Context, *pb.Filter, func(laptop *pb.Laptop) error) error
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

func (i *InMemoryLaptopStore) Search(ctx context.Context, filter *pb.Filter,
	found func(laptop *pb.Laptop) error) error {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	for _, laptop := range i.data {

		log.Print("checking laptop id", laptop.Id)
		if ctx.Err() != nil {
			log.Printf("context cancel with err %v", ctx.Err())
		}

		if isQualified(filter, laptop) {
			err := found(laptop)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func isQualified(filter *pb.Filter, laptop *pb.Laptop) bool {
	if laptop.GetPriceUsd() > filter.GetMaxPriceUsd() {
		return false
	}

	if laptop.GetCpu().GetNumberCores() < filter.GetMinCpuCores() {
		return false
	}

	if laptop.GetCpu().GetMinGhz() < filter.GetMinCpuGhz() {
		return false
	}

	if toBit(laptop.GetRam()) < toBit(filter.MinRam) {
		return false
	}

	return true
}

func toBit(memory *pb.Memory) uint64 {
	val := memory.GetValue()

	switch memory.GetUnit() {
	case pb.Memory_BIT:
		return val
	case pb.Memory_BYTE:
		return val << 3
	case pb.Memory_KILOBYTE:
		return val << 13
	case pb.Memory_MEGABYTE:
		return val << 23
	case pb.Memory_GIGABYTE:
		return val << 33
	case pb.Memory_TERABYTE:
		return val << 43
	default:
		return 0
	}
}
