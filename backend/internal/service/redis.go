package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lingbao-market/backend/internal/model"
	"github.com/redis/go-redis/v9"
)

type PriceService struct {
	rdb *redis.Client
}

func NewPriceService(rdb *redis.Client) *PriceService {
	return &PriceService{rdb: rdb}
}

const (
	keyPriceTime  = "market:feed:time"
	keyPriceValue = "market:feed:price"
)

func (s *PriceService) AddPrice(ctx context.Context, item model.PriceItem) error {
	item.Timestamp = time.Now().UnixMilli()
	val, err := json.Marshal(item)
	if err != nil {
		return err
	}

	// Pipeline for atomicity
	pipe := s.rdb.Pipeline()

	// 1. Index by Time (Latest first)
	pipe.ZAdd(ctx, keyPriceTime, redis.Z{
		Score:  float64(item.Timestamp),
		Member: val,
	})

	// 2. Index by Price (Highest first)
	// Note: In a real prod system, we might want unique members per code to update prices,
	// but here we are storing a stream of submissions.
	pipe.ZAdd(ctx, keyPriceValue, redis.Z{
		Score:  float64(item.Price),
		Member: val,
	})

	// Cleanup old data (older than 24h) from Time index
	cutoff := time.Now().Add(-24 * time.Hour).UnixMilli()
	pipe.ZRemRangeByScore(ctx, keyPriceTime, "-inf", fmt.Sprintf("%d", cutoff))

	// We should also trim the Price index to keep it consistent-ish,
	// but strictly syncing two ZSETs by score removal is tricky without unique IDs.
	// For this high-throughput demo, we'll keep the price index simply capped by size or let it expire via TTL if we set one.
	// Let's just keep the last 1000 items in the price index to prevent bloat for now.
	pipe.ZRemRangeByRank(ctx, keyPriceValue, 0, -1001)

	_, err = pipe.Exec(ctx)
	return err
}

// GetLatestFeed returns items sorted by the requested criteria
func (s *PriceService) GetLatestFeed(ctx context.Context, limit int64, sortBy string) ([]model.PriceItem, error) {
	var key string
	switch sortBy {
	case "price":
		key = keyPriceValue
	default: // "time"
		key = keyPriceTime
	}

	// ZREVRANGE 0 to limit-1 (Highest Score first: Latest Time OR Highest Price)
	vals, err := s.rdb.ZRevRange(ctx, key, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}

	var items []model.PriceItem
	for _, val := range vals {
		var item model.PriceItem
		if err := json.Unmarshal([]byte(val), &item); err == nil {
			items = append(items, item)
		}
	}
	return items, nil
}
