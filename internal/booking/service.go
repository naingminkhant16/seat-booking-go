package booking

type Service struct {
	store BookingStore
}

func NewService(s BookingStore) *Service {
	return &Service{s}
}

func (s *Service) Book(b Booking) error {
	err := s.store.Book(b)
	return err
}
