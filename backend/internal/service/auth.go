package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/lingbao-market/backend/internal/model"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	rdb       *redis.Client
	jwtSecret []byte
}

const (
	userKeyPrefix = "auth:user:"
	userIndexKey  = "auth:users"
)

func NewAuthService(rdb *redis.Client, jwtSecret string) *AuthService {
	return &AuthService{
		rdb:       rdb,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *AuthService) Register(ctx context.Context, username, password string) (*model.User, error) {
	return s.CreateUser(ctx, username, password, false)
}

func (s *AuthService) CreateUser(ctx context.Context, username, password string, isAdmin bool) (*model.User, error) {
	exists, err := s.rdb.Exists(ctx, userKeyPrefix+username).Result()
	if err != nil {
		return nil, err
	}
	if exists > 0 {
		return nil, errors.New("username already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := model.User{
		ID:           uuid.New().String(),
		Username:     username,
		PasswordHash: string(hash),
		IsAdmin:      isAdmin,
		Banned:       false,
	}

	if err := s.saveUser(ctx, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*model.AuthResponse, error) {
	// 1. Get User
	val, err := s.rdb.Get(ctx, userKeyPrefix+username).Result()
	if err == redis.Nil {
		return nil, errors.New("invalid credentials")
	}
	if err != nil {
		return nil, err
	}

	var user model.User
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return nil, err
	}
	if user.Banned {
		return nil, errors.New("account banned")
	}

	// 2. Verify Password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// 3. Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      user.ID,
		"username": user.Username,
		"admin":    user.IsAdmin,
		"exp":      time.Now().Add(72 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		Token:    tokenString,
		Username: user.Username,
		ID:       user.ID,
		IsAdmin:  user.IsAdmin,
	}, nil
}

func (s *AuthService) CreateCaptcha(ctx context.Context) (string, string, error) {
	id := uuid.New().String()
	code, err := randomCode(4)
	if err != nil {
		return "", "", err
	}

	if err := s.rdb.Set(ctx, "auth:captcha:"+id, code, 5*time.Minute).Err(); err != nil {
		return "", "", err
	}

	return id, code, nil
}

func (s *AuthService) VerifyCaptcha(ctx context.Context, id, code string) (bool, error) {
	if id == "" || code == "" {
		return false, nil
	}

	val, err := s.rdb.Get(ctx, "auth:captcha:"+id).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return strings.EqualFold(val, code), nil
}

func (s *AuthService) DeleteCaptcha(ctx context.Context, id string) error {
	if id == "" {
		return nil
	}
	return s.rdb.Del(ctx, "auth:captcha:"+id).Err()
}

func (s *AuthService) EnsureAdmin(ctx context.Context, username, password string) (*model.User, error) {
	if username == "" || password == "" {
		return nil, nil
	}
	exists, err := s.rdb.Exists(ctx, userKeyPrefix+username).Result()
	if err != nil {
		return nil, err
	}
	if exists > 0 {
		user, err := s.GetUser(ctx, username)
		if err != nil {
			return nil, err
		}
		updated := false
		if !user.IsAdmin {
			user.IsAdmin = true
			updated = true
		}
		if password != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return nil, err
			}
			user.PasswordHash = string(hash)
			updated = true
		}
		if updated {
			if err := s.saveUser(ctx, user); err != nil {
				return nil, err
			}
		}
		return user, nil
	}

	return s.CreateUser(ctx, username, password, true)
}

func (s *AuthService) GetUser(ctx context.Context, username string) (*model.User, error) {
	val, err := s.rdb.Get(ctx, userKeyPrefix+username).Result()
	if err == redis.Nil {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	var user model.User
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *AuthService) ListUsers(ctx context.Context) ([]model.User, error) {
	usernames, err := s.rdb.SMembers(ctx, userIndexKey).Result()
	if err != nil {
		return nil, err
	}

	if len(usernames) == 0 {
		var cursor uint64
		for {
			keys, nextCursor, err := s.rdb.Scan(ctx, cursor, userKeyPrefix+"*", 200).Result()
			if err != nil {
				return nil, err
			}
			for _, key := range keys {
				username := strings.TrimPrefix(key, userKeyPrefix)
				if username != "" {
					usernames = append(usernames, username)
				}
			}
			cursor = nextCursor
			if cursor == 0 {
				break
			}
		}
	}

	var users []model.User
	for _, username := range usernames {
		user, err := s.GetUser(ctx, username)
		if err == nil {
			users = append(users, *user)
		}
	}
	return users, nil
}

func (s *AuthService) SetBanned(ctx context.Context, username string, banned bool) (*model.User, error) {
	user, err := s.GetUser(ctx, username)
	if err != nil {
		return nil, err
	}
	user.Banned = banned
	if err := s.saveUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) DeleteUser(ctx context.Context, username string) error {
	if err := s.rdb.Del(ctx, userKeyPrefix+username).Err(); err != nil {
		return err
	}
	return s.rdb.SRem(ctx, userIndexKey, username).Err()
}

func (s *AuthService) IsBanned(ctx context.Context, username string) (bool, error) {
	user, err := s.GetUser(ctx, username)
	if err != nil {
		return false, err
	}
	return user.Banned, nil
}

func (s *AuthService) saveUser(ctx context.Context, user *model.User) error {
	val, err := json.Marshal(user)
	if err != nil {
		return err
	}

	pipe := s.rdb.Pipeline()
	pipe.Set(ctx, userKeyPrefix+user.Username, val, 0)
	pipe.SAdd(ctx, userIndexKey, user.Username)
	_, err = pipe.Exec(ctx)
	return err
}

func randomCode(length int) (string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[n.Int64()]
	}
	return string(result), nil
}
