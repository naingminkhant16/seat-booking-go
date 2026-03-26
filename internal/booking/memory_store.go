package booking

type MemoryStore struct {
	bookings map[string]Booking
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{bookings: make(map[string]Booking)}
}

func (s *MemoryStore) Book(b Booking) error {
	// check if it reserved
	if _, exists := s.bookings[b.SeatID]; exists {
		return ErrSeatAlreadyTaken
	}
	// make booking reserved
	s.bookings[b.SeatID] = b
	return nil
}

func (s *MemoryStore) ListBookings(movieID string) []Booking {
	var bookings []Booking
	for _, b := range s.bookings {
		if b.MovieID == movieID {
			bookings = append(bookings, b)
		}
	}
	return bookings
}
