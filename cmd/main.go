package main

import (
	"MovieSeatBooking/internal/booking"
	"log"
)

func main() {
	newStore := booking.NewMemoryStore()
	service := booking.NewService(newStore)

	err := service.Book(booking.Booking{
		ID:      "1",
		MovieID: "1",
		SeatID:  "A1",
		UserID:  "10",
		Status:  "reserved",
	})
	if err != nil {
		return
	}
	log.Println(newStore.ListBookings("1"))
}
