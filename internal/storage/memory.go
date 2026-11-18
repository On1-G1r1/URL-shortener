package storage

import (
	"errors"
	"sync"
)

var ErrNotFound = errors.New("key not found")

type MemoryStorage struct{
	mu   sync.RWMutex
	data map[string]string
	
}
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string]string),
	}
}
func (s *MemoryStorage) Set(key, value string) {

	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}
func (s *MemoryStorage) Get(key string) (string, error) {

s.mu.RLock()
	defer s.mu.RUnlock()

	val, exists := s.data[key]
	if !exists {
		return "", ErrNotFound
	}
	return val, nil
}
func (s *MemoryStorage) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key)
}
func (s *MemoryStorage) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.data)
}
