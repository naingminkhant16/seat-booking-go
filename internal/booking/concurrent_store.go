package booking

import "sync"

type ConcurrentStore struct {
	bookings map[string]Booking
	sync.RWMutex
}

func NewConcurrentStore() *ConcurrentStore {
	return &ConcurrentStore{bookings: make(map[string]Booking)}
}

func (s *ConcurrentStore) Book(b Booking) error {
	// This prevents double booking
	s.Lock()
	defer s.Unlock()

	// check if it reserved
	if _, exists := s.bookings[b.SeatID]; exists {
		return ErrSeatAlreadyTaken
	}
	// make booking reserved
	s.bookings[b.SeatID] = b
	return nil
}

func (s *ConcurrentStore) ListBookings(movieID string) []Booking {
	s.RLock()
	defer s.RUnlock()

	var bookings []Booking
	for _, b := range s.bookings {
		if b.MovieID == movieID {
			bookings = append(bookings, b)
		}
	}
	return bookings
}
