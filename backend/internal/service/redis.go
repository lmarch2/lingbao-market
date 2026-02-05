package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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

func (s *PriceService) ClearAllPrices(ctx context.Context) (int64, int64, error) {
	pipe := s.rdb.TxPipeline()
	timeCount := pipe.ZCard(ctx, keyPriceTime)
	priceCount := pipe.ZCard(ctx, keyPriceValue)
	pipe.Del(ctx, keyPriceTime, keyPriceValue)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, 0, err
	}

	return timeCount.Val(), priceCount.Val(), nil
}

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

	_, err = pipe.Exec(ctx)
	return err
}

func (s *PriceService) CleanupExpired(ctx context.Context, cutoff time.Time) (int64, int64, error) {
	cutoffMillis := cutoff.UnixMilli()

	removedTime, err := s.rdb.ZRemRangeByScore(ctx, keyPriceTime, "-inf", fmt.Sprintf("%d", cutoffMillis)).Result()
	if err != nil {
		return 0, 0, err
	}

	var removedPrice int64
	var cursor uint64
	for {
		vals, nextCursor, err := s.rdb.ZScan(ctx, keyPriceValue, cursor, "*", 200).Result()
		if err != nil {
			return removedTime, removedPrice, err
		}

		var toRemove []interface{}
		for i := 0; i < len(vals); i += 2 {
			member := vals[i]
			var item model.PriceItem
			if err := json.Unmarshal([]byte(member), &item); err != nil {
				continue
			}
			if item.Timestamp > 0 && item.Timestamp < cutoffMillis {
				toRemove = append(toRemove, member)
			}
		}

		if len(toRemove) > 0 {
			removed, err := s.rdb.ZRem(ctx, keyPriceValue, toRemove...).Result()
			if err != nil {
				return removedTime, removedPrice, err
			}
			removedPrice += removed
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return removedTime, removedPrice, nil
}

func (s *PriceService) DeletePricesByCode(ctx context.Context, code string) (int64, int64, error) {
	removedTime, err := s.removeByCode(ctx, keyPriceTime, code)
	if err != nil {
		return 0, 0, err
	}
	removedPrice, err := s.removeByCode(ctx, keyPriceValue, code)
	if err != nil {
		return removedTime, 0, err
	}
	return removedTime, removedPrice, nil
}

func (s *PriceService) removeByCode(ctx context.Context, key, code string) (int64, error) {
	var removedTotal int64
	var cursor uint64
	for {
		vals, nextCursor, err := s.rdb.ZScan(ctx, key, cursor, "*", 200).Result()
		if err != nil {
			return removedTotal, err
		}

		var toRemove []interface{}
		for i := 0; i < len(vals); i += 2 {
			member := vals[i]
			var item model.PriceItem
			if err := json.Unmarshal([]byte(member), &item); err != nil {
				continue
			}
			if strings.EqualFold(item.Code, code) {
				toRemove = append(toRemove, member)
			}
		}

		if len(toRemove) > 0 {
			removed, err := s.rdb.ZRem(ctx, key, toRemove...).Result()
			if err != nil {
				return removedTotal, err
			}
			removedTotal += removed
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return removedTotal, nil
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
