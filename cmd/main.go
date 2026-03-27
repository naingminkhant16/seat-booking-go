package main

import (
	"MovieSeatBooking/internal/adapters/redis"
	"MovieSeatBooking/internal/booking"
	"MovieSeatBooking/internal/utils"
	"log"
	"net/http"
)

type movieResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Rows        int    `json:"rows"`
	SeatsPerRow int    `json:"seats_per_row"`
}

func main() {
	mux := http.NewServeMux()

	redisStore := booking.NewRedisStore(redis.NewClient("localhost:6379"))
	svc := booking.NewService(redisStore)
	bookingHandler := booking.NewHandler(svc)

	mux.HandleFunc("GET /movies", listMovies)
	mux.Handle("GET /", http.FileServer(http.Dir("static")))

	mux.HandleFunc("GET /movies/{movieID}/seats", bookingHandler.ListSeats)
	mux.HandleFunc("POST /movies/{movieID}/seats/{seatID}/hold", bookingHandler.SeatHold)

	mux.HandleFunc("PUT /sessions/{sessionID}/confirm", bookingHandler.ConfirmSession)
	mux.HandleFunc("DELETE /sessions/{sessionID}", bookingHandler.ReleaseSession)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}

}

var movies = []movieResponse{
	{ID: "inception", Title: "Inception", Rows: 5, SeatsPerRow: 8},
	{ID: "dune", Title: "Dune: Part Two", Rows: 4, SeatsPerRow: 6},
}

func listMovies(w http.ResponseWriter, req *http.Request) {
	utils.WriteJSON(w, http.StatusOK, movies)
}
