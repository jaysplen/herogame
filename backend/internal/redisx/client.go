package redisx

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const arrivalsKey = "arrivals:zset"

// Client wraps Redis for transient game state.
type Client struct {
	rdb *redis.Client
}

// New parses redisURL (e.g. redis://localhost:6379/0) and connects.
func New(ctx context.Context, redisURL string) (*Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	rdb := redis.NewClient(opts)
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &Client{rdb: rdb}, nil
}

// NewFromClient wraps an existing go-redis client (tests).
func NewFromClient(rdb *redis.Client) *Client {
	return &Client{rdb: rdb}
}

// Close closes the connection.
func (c *Client) Close() error {
	return c.rdb.Close()
}

// Underlying exposes the raw client.
func (c *Client) Underlying() *redis.Client {
	return c.rdb
}

// AddArrival schedules a movement order for resolution at arriveAt.
func (c *Client) AddArrival(ctx context.Context, orderID int64, arriveAt time.Time) error {
	score := float64(arriveAt.Unix())
	member := strconv.FormatInt(orderID, 10)
	return c.rdb.ZAdd(ctx, arrivalsKey, redis.Z{Score: score, Member: member}).Err()
}

// DueArrivals returns movement order IDs with arrive_at <= now, up to limit.
func (c *Client) DueArrivals(ctx context.Context, now time.Time, limit int64) ([]int64, error) {
	max := strconv.FormatInt(now.Unix(), 10)
	members, err := c.rdb.ZRangeByScore(ctx, arrivalsKey, &redis.ZRangeBy{
		Min:   "0",
		Max:   max,
		Count: limit,
	}).Result()
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(members))
	for _, m := range members {
		id, err := strconv.ParseInt(m, 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// RemoveArrival removes an order from the arrivals index.
func (c *Client) RemoveArrival(ctx context.Context, orderID int64) error {
	return c.rdb.ZRem(ctx, arrivalsKey, strconv.FormatInt(orderID, 10)).Err()
}

// Rehydrate rebuilds the ZSET from in-flight DB rows (boot recovery).
func (c *Client) Rehydrate(ctx context.Context, entries []ArrivalEntry) error {
	if len(entries) == 0 {
		return nil
	}
	pipe := c.rdb.Pipeline()
	pipe.Del(ctx, arrivalsKey)
	for _, e := range entries {
		pipe.ZAdd(ctx, arrivalsKey, redis.Z{
			Score: float64(e.ArriveAt.Unix()),
			Member: strconv.FormatInt(e.OrderID, 10),
		})
	}
	_, err := pipe.Exec(ctx)
	return err
}

// ArrivalEntry is one in-flight movement for ZSET rebuild.
type ArrivalEntry struct {
	OrderID  int64
	ArriveAt time.Time
}
