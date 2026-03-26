package booking

import "errors"

var (
	ErrSeatAlreadyTaken = errors.New("seat already taken")
)

// Booking Representation of confirmed seat reservation
type Booking struct {
	ID      string
	MovieID string
	SeatID  string
	UserID  string
	Status  string
}

type BookingStore interface {
	Book(b Booking) error
	ListBookings(movieID string) []Booking
}
