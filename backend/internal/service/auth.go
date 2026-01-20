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

func NewAuthService(rdb *redis.Client, jwtSecret string) *AuthService {
	return &AuthService{
		rdb:       rdb,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *AuthService) Register(ctx context.Context, username, password string) (*model.User, error) {
	// 1. Check if user exists
	exists, err := s.rdb.Exists(ctx, "auth:user:"+username).Result()
	if err != nil {
		return nil, err
	}
	if exists > 0 {
		return nil, errors.New("username already exists")
	}

	// 2. Hash Password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 3. Create User
	user := model.User{
		ID:           uuid.New().String(),
		Username:     username,
		PasswordHash: string(hash),
	}

	// 4. Save to Redis
	val, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	if err := s.rdb.Set(ctx, "auth:user:"+username, val, 0).Err(); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*model.AuthResponse, error) {
	// 1. Get User
	val, err := s.rdb.Get(ctx, "auth:user:"+username).Result()
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

	// 2. Verify Password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// 3. Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"name": user.Username,
		"exp":  time.Now().Add(72 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		Token:    tokenString,
		Username: user.Username,
		ID:       user.ID,
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
