package future

import (
	"sync"
	"time"

	"github.com/256dpi/gomqtt/packet"
)

// A Store is used to store futures.
type Store struct {
	sync.RWMutex

	protected bool
	store     map[packet.ID]*Future
}

// NewStore will create a new Store.
func NewStore() *Store {
	return &Store{
		store: make(map[packet.ID]*Future),
	}
}

// Put will save a future to the store.
func (s *Store) Put(id packet.ID, future *Future) {
	s.Lock()
	defer s.Unlock()

	s.store[id] = future
}

// Get will retrieve a future from the store.
func (s *Store) Get(id packet.ID) *Future {
	s.RLock()
	defer s.RUnlock()

	return s.store[id]
}

// Delete will remove a future from the store.
func (s *Store) Delete(id packet.ID) {
	s.Lock()
	defer s.Unlock()

	delete(s.store, id)
}

// All will return a slice with all stored futures.
func (s *Store) All() []*Future {
	s.RLock()
	defer s.RUnlock()

	all := make([]*Future, 0, len(s.store))

	for _, savedFuture := range s.store {
		all = append(all, savedFuture)
	}

	return all
}

// Protect will set the protection attribute and if true prevents the store from
// being cleared.
func (s *Store) Protect(value bool) {
	s.Lock()
	defer s.Unlock()

	s.protected = value
}

// Clear will cancel all stored futures and remove them if the store is unprotected.
func (s *Store) Clear() {
	s.Lock()
	defer s.Unlock()

	if s.protected {
		return
	}

	for _, savedFuture := range s.store {
		savedFuture.Cancel()
	}

	s.store = make(map[packet.ID]*Future)
}

// Await will wait until all futures have completed and removed or timeout is
// reached.
func (s *Store) Await(timeout time.Duration) error {
	stop := time.Now().Add(timeout)

	for {
		// Get futures
		s.RLock()
		futures := s.All()
		s.RUnlock()

		// return if no futures are left
		if len(futures) == 0 {
			return nil
		}

		// wait for next future to complete
		err := futures[0].Wait(stop.Sub(time.Now()))
		if err != nil {
			return err
		}
	}
}
