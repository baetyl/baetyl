package utils

import "sync"

type Set struct {
	m      map[interface{}]bool
	length int
	sync.RWMutex
}

func NewSet() *Set {
	return &Set{
		m:      map[interface{}]bool{},
		length: 0,
	}
}

func (s *Set) Add(item interface{}) {
	s.Lock()
	defer s.Unlock()
	if !s.m[item] {
		s.m[item] = true
		s.length++
	}
}

func (s *Set) Has(item interface{}) bool {
	s.RLock()
	defer s.RUnlock()
	_, ok := s.m[item]
	return ok
}

func (s *Set) Len() int {
	return s.length
}

func (s *Set) IsEmpty() bool {
	return s.length == 0
}

func (s *Set) List() []interface{} {
	s.RLock()
	defer s.RUnlock()
	list := make([]interface{}, 0)
	for item, _ := range s.m {
		list = append(list, item)
	}
	return list
}
