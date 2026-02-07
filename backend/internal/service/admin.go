package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lingbao-market/backend/internal/model"
	"github.com/redis/go-redis/v9"
)

type AdminService struct {
	rdb *redis.Client
}

const (
	feedbackKeyPrefix = "admin:feedback:"
	feedbackIndexKey  = "admin:feedback:index"
	adminLogsKey      = "admin:logs"
	adminLogsMax      = 500
)

func NewAdminService(rdb *redis.Client) *AdminService {
	return &AdminService{rdb: rdb}
}

func (s *AdminService) AddFeedback(ctx context.Context, code, reason, reporter string) (*model.FeedbackMessage, error) {
	entry := model.FeedbackMessage{
		ID:        uuid.New().String(),
		Code:      code,
		Reason:    reason,
		Reporter:  reporter,
		CreatedAt: time.Now().UnixMilli(),
		Resolved:  false,
	}

	val, err := json.Marshal(entry)
	if err != nil {
		return nil, err
	}

	pipe := s.rdb.TxPipeline()
	pipe.Set(ctx, feedbackKeyPrefix+entry.ID, val, 0)
	pipe.ZAdd(ctx, feedbackIndexKey, redis.Z{
		Score:  float64(entry.CreatedAt),
		Member: entry.ID,
	})

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	err = s.AppendLog(ctx, model.AdminLogEntry{
		Type:      "feedback_submitted",
		Message:   fmt.Sprintf("feedback submitted for code %s", code),
		Actor:     reporter,
		Timestamp: entry.CreatedAt,
		Metadata: map[string]string{
			"feedbackId": entry.ID,
			"code":       code,
		},
	})
	if err != nil {
		return nil, err
	}

	return &entry, nil
}

func (s *AdminService) GetFeedback(ctx context.Context, id string) (*model.FeedbackMessage, error) {
	val, err := s.rdb.Get(ctx, feedbackKeyPrefix+id).Result()
	if err == redis.Nil {
		return nil, errors.New("feedback not found")
	}
	if err != nil {
		return nil, err
	}

	var feedback model.FeedbackMessage
	if err := json.Unmarshal([]byte(val), &feedback); err != nil {
		return nil, err
	}
	return &feedback, nil
}

func (s *AdminService) ListFeedback(ctx context.Context, limit int64, includeResolved bool) ([]model.FeedbackMessage, error) {
	if limit <= 0 {
		limit = 100
	}

	ids, err := s.rdb.ZRevRange(ctx, feedbackIndexKey, 0, 499).Result()
	if err != nil {
		return nil, err
	}

	messages := make([]model.FeedbackMessage, 0, limit)
	for _, id := range ids {
		item, err := s.GetFeedback(ctx, id)
		if err != nil {
			continue
		}
		if !includeResolved && item.Resolved {
			continue
		}
		messages = append(messages, *item)
		if int64(len(messages)) >= limit {
			break
		}
	}

	return messages, nil
}

func (s *AdminService) ResolveFeedback(
	ctx context.Context,
	id, resolver, action string,
	removedTime, removedPrice int64,
) (*model.FeedbackMessage, error) {
	feedback, err := s.GetFeedback(ctx, id)
	if err != nil {
		return nil, err
	}
	if feedback.Resolved {
		return nil, errors.New("feedback already resolved")
	}

	now := time.Now().UnixMilli()
	feedback.Resolved = true
	feedback.ResolvedAt = now
	feedback.ResolvedBy = resolver
	feedback.Action = action
	feedback.RemovedTime = removedTime
	feedback.RemovedPrice = removedPrice

	val, err := json.Marshal(feedback)
	if err != nil {
		return nil, err
	}

	if err := s.rdb.Set(ctx, feedbackKeyPrefix+feedback.ID, val, 0).Err(); err != nil {
		return nil, err
	}

	message := fmt.Sprintf("feedback %s resolved by %s", feedback.ID, resolver)
	if action == "delete" {
		message = fmt.Sprintf(
			"feedback %s resolved with delete on code %s",
			feedback.ID,
			feedback.Code,
		)
	}

	err = s.AppendLog(ctx, model.AdminLogEntry{
		Type:      "feedback_resolved",
		Message:   message,
		Actor:     resolver,
		Timestamp: now,
		Metadata: map[string]string{
			"feedbackId":   feedback.ID,
			"code":         feedback.Code,
			"action":       action,
			"removedTime":  fmt.Sprintf("%d", removedTime),
			"removedPrice": fmt.Sprintf("%d", removedPrice),
		},
	})
	if err != nil {
		return nil, err
	}

	return feedback, nil
}

func (s *AdminService) AppendLog(ctx context.Context, entry model.AdminLogEntry) error {
	if strings.TrimSpace(entry.ID) == "" {
		entry.ID = uuid.New().String()
	}
	if entry.Timestamp <= 0 {
		entry.Timestamp = time.Now().UnixMilli()
	}
	if strings.TrimSpace(entry.Actor) == "" {
		entry.Actor = "system"
	}

	val, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	pipe := s.rdb.TxPipeline()
	pipe.LPush(ctx, adminLogsKey, val)
	pipe.LTrim(ctx, adminLogsKey, 0, adminLogsMax-1)
	_, err = pipe.Exec(ctx)
	return err
}

func (s *AdminService) ListLogs(ctx context.Context, limit int64) ([]model.AdminLogEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > adminLogsMax {
		limit = adminLogsMax
	}

	vals, err := s.rdb.LRange(ctx, adminLogsKey, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}

	entries := make([]model.AdminLogEntry, 0, len(vals))
	for _, val := range vals {
		var entry model.AdminLogEntry
		if err := json.Unmarshal([]byte(val), &entry); err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	return entries, nil
}
