package storage

import (
	"sync"
)

// -------------------------------------------------

func NewLocalStorage() *keyValueStorage {
	return newKeyValueStorage(&localStorageDriver{
		// Lock
		Lock: &sync.Mutex{},

		// Table
		Users: make(map[string]string),
	})
}

type localStorageDriver struct {
	Lock *sync.Mutex

	// Map{userID => userJson}
	Users map[string]string
}

func (s *localStorageDriver) Set(userID, userJson string) error {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	s.Users[userID] = userJson
	return nil
}

func (s *localStorageDriver) Lookup(userID string) (string, bool, error) {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	userJson, ok := s.Users[userID]
	return userJson, ok, nil
}

func (s *localStorageDriver) Index(name string) keyValueIndex {
	return NewIndex()
}

// -------------------------------------------------

func NewIndex() *Index {
	return &Index{
		&sync.Mutex{},
		make(map[string]string),
	}
}

// Index is a helper struct
type Index struct {
	Lock *sync.Mutex
	Data map[string]string
}

func (i *Index) Put(key, value string) error {
	i.Lock.Lock()
	defer i.Lock.Unlock()

	i.Data[key] = value
	return nil
}

func (i *Index) Remove(key string) error {
	i.Lock.Lock()
	defer i.Lock.Unlock()

	delete(i.Data, key)
	return nil
}

func (i *Index) Lookup(key string) (string, bool, error) {
	i.Lock.Lock()
	defer i.Lock.Unlock()

	value, ok := i.Data[key]
	return value, ok, nil
}
