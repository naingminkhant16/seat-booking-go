package booking

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const defaultHoldTTL = 2 * time.Minute

// RedisStore implements session-based seat booking with Redis
// seat:{movieID}:{seatID} -> session JSON (TTL = held, No TTL = confirmed)
// session:{sessionID} -> seat key (reverse lookup)
type RedisStore struct {
	rdb *redis.Client
}

func NewRedisStore(rdb *redis.Client) *RedisStore {
	return &RedisStore{rdb: rdb}
}

// sessionKey builds the reserve lookup key for session
func sessionKey(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}
func parseSession(val string) (Booking, error) {
	var data Booking

	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return Booking{}, err
	}

	return Booking{
		ID:      data.ID,
		UserID:  data.UserID,
		MovieID: data.MovieID,
		SeatID:  data.SeatID,
		Status:  data.Status,
	}, nil
}
func (s *RedisStore) Book(b Booking) (Booking, error) {
	session, err := s.hold(b)
	if err != nil {
		return Booking{}, err
	}
	log.Printf("Session booked %v", session)

	return session, nil
}
func (s *RedisStore) ListBookings(movieID string) []Booking {
	pattern := fmt.Sprintf("seat:%s:*", movieID)
	var sessions []Booking
	ctx := context.Background()

	iter := s.rdb.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		val, err := s.rdb.Get(ctx, iter.Val()).Result()
		if err != nil {
			continue
		}
		session, err := parseSession(val)
		if err != nil {
			continue
		}
		sessions = append(sessions, session)
	}
	return sessions
}

func (s *RedisStore) hold(b Booking) (Booking, error) {
	id := uuid.New().String()
	b.ID = id

	now := time.Now()
	ctx := context.Background()
	key := fmt.Sprintf("seat:%s:%s", b.MovieID, b.SeatID)
	val, _ := json.Marshal(b)
	args := redis.SetArgs{
		Mode: "NX", // set only key doesn't exist
		TTL:  defaultHoldTTL,
	}
	// store booking
	res := s.rdb.SetArgs(ctx, key, val, args)
	if ok := res.Val() == "OK"; !ok {
		return Booking{}, ErrSeatAlreadyTaken
	}

	// store booking key with session id
	s.rdb.Set(ctx, sessionKey(id), key, defaultHoldTTL)

	return Booking{
		ID:        id,
		MovieID:   b.MovieID,
		SeatID:    b.SeatID,
		UserID:    b.UserID,
		Status:    "held",
		ExpiresAt: now.Add(defaultHoldTTL),
	}, nil
}

func (s *RedisStore) Confirm(ctx context.Context, sessionID string, userID string) (Booking, error) {
	session, sk, err := s.getSession(ctx, sessionID, userID)
	if err != nil {
		return Booking{}, err
	}

	s.rdb.Persist(ctx, sk) // Remove TTL
	s.rdb.Persist(ctx, sessionKey(sessionID))

	session.Status = "confirmed"
	data := Booking{
		ID:      sessionID,
		UserID:  userID,
		MovieID: session.MovieID,
		SeatID:  session.SeatID,
		Status:  "confirmed",
	}
	val, _ := json.Marshal(data)
	s.rdb.Set(ctx, sk, val, 0)
	return session, nil
}

func (s *RedisStore) Release(ctx context.Context, sessionID string, userID string) error {
	_, sk, err := s.getSession(ctx, sessionID, userID)
	if err != nil {
		return err
	}
	s.rdb.Del(ctx, sk, sessionKey(sessionID))
	return nil
}

func (s *RedisStore) getSession(ctx context.Context, sessionID string, userID string) (Booking, string, error) {
	sk, err := s.rdb.Get(ctx, sessionKey(sessionID)).Result()
	if err != nil {
		return Booking{}, "", err
	}

	val, err := s.rdb.Get(ctx, sk).Result()
	if err != nil {
		return Booking{}, "", err
	}

	session, err := parseSession(val)
	if err != nil {
		return Booking{}, "", err
	}

	return session, sk, nil
}
