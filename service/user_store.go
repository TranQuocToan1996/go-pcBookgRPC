package service

import (
	"fmt"
	"sync"
)

type UserStore interface {
	Save(user *User) error
	Find(username string) (*User, error)
}

type InMemoryUserStore struct {
	users map[string]*User
	mutex sync.RWMutex
}

func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		users: make(map[string]*User),
	}
}

func (store *InMemoryUserStore) Save(user *User) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	found := store.users[user.UserName]
	if found != nil {
		return ErrAlreadyExist
	}

	store.users[user.UserName] = user.Clone()

	return nil
}

func (store *InMemoryUserStore) Find(username string) (*User, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	found := store.users[username]
	if found != nil {
		return found.Clone(), nil
	}

	return nil, fmt.Errorf("cant found user in store")
}
