package utils

import "sync"

// Set unsorted set, implement with map
type Set struct {
	m      map[interface{}]bool
	length int
	sync.RWMutex
}

// NewSet create a new set, return the pointer
func NewSet() *Set {
	return &Set{
		m:      map[interface{}]bool{},
		length: 0,
	}
}

// Add add element to the set
func (s *Set) Add(item interface{}) {
	s.Lock()
	defer s.Unlock()
	if !s.m[item] {
		s.m[item] = true
		s.length++
	}
}

// Remove remove a specify element in the set
func (s *Set) Remove(item interface{}) {
	s.Lock()
	defer s.Unlock()
	delete(s.m, item)
}

// Has judge a set has a element or not
func (s *Set) Has(item interface{}) bool {
	s.RLock()
	defer s.RUnlock()
	_, ok := s.m[item]
	return ok
}

// Len remove the length of set
func (s *Set) Len() int {
	return s.length
}

// IsEmpty set is empty or not
func (s *Set) IsEmpty() bool {
	return s.length == 0
}

// List transfer the set to slice
func (s *Set) List() []interface{} {
	s.RLock()
	defer s.RUnlock()
	list := make([]interface{}, 0)
	for item, _ := range s.m {
		list = append(list, item)
	}
	return list
}
