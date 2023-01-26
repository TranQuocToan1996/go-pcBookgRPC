package service

import "sync"

type RatingStore interface {
	Add(laptopID string, score float64) (*Rating, error)
}

type Rating struct {
	Count uint32
	Sum   float64
}

type InMemoryRatingStore struct {
	m      sync.RWMutex
	rating map[string]*Rating
}

func (store *InMemoryRatingStore) Add(laptopID string, score float64) (*Rating, error) {
	store.m.Lock()
	defer store.m.Unlock()

	exist, ok := store.rating[laptopID]
	if ok {
		rating := exist
		rating.Count += 1
		rating.Sum += score
		store.rating[laptopID] = rating
	} else {
		store.rating[laptopID] = &Rating{Count: 1, Sum: score}
	}

	return store.rating[laptopID], nil
}

func NewInMemoryRatingStore() RatingStore {
	return &InMemoryRatingStore{
		rating: make(map[string]*Rating),
	}
}
