package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/lingbao-market/backend/internal/api"
	"github.com/lingbao-market/backend/internal/config"
	"github.com/lingbao-market/backend/internal/service"
	"github.com/redis/go-redis/v9"
)

func main() {
	// 1. Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Init Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})

	// Check Redis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}

	// 3. Init Services
	svc := service.NewPriceService(rdb)
	authSvc := service.NewAuthService(rdb, cfg.JWTSecret)

	if _, err := authSvc.EnsureAdmin(context.Background(), cfg.AdminUsername, cfg.AdminPassword); err != nil {
		log.Printf("Failed to ensure admin user: %v", err)
	}

	// 4. Init Fiber
	app := fiber.New(fiber.Config{
		AppName: "Lingbao Market Backend",
	})

	// 4. Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(limiter.New(limiter.Config{
		Max:        10,
		Expiration: 30 * time.Second,
		Next: func(c *fiber.Ctx) bool {
			if c.Method() == fiber.MethodOptions {
				return true
			}
			if c.Method() == fiber.MethodGet && c.Path() == "/api/v1/feed" {
				return true
			}
			return false
		},
		KeyGenerator: func(c *fiber.Ctx) string {
			xff := strings.TrimSpace(c.Get(fiber.HeaderXForwardedFor))
			if xff != "" {
				parts := strings.Split(xff, ",")
				if len(parts) > 0 {
					ip := strings.TrimSpace(parts[0])
					if ip != "" {
						return ip
					}
				}
			}
			xri := strings.TrimSpace(c.Get("X-Real-IP"))
			if xri != "" {
				return xri
			}
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many requests, please try again later.",
			})
		},
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // Lock down in production
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
	}))

	// 5. Routes
	h := api.NewHandler(svc, authSvc, cfg.JWTSecret)
	h.RegisterRoutes(app)

	// 6. Schedule daily cleanup
	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	defer cleanupCancel()
	cleanupHour, cleanupMinute, err := parseCleanupTime(cfg.CleanupTime)
	if err != nil {
		log.Printf("Invalid CLEANUP_TIME %q, falling back to 03:00", cfg.CleanupTime)
		cleanupHour, cleanupMinute = 3, 0
	}
	loc, err := time.LoadLocation(cfg.CleanupTimezone)
	if err != nil {
		log.Printf("Invalid CLEANUP_TIMEZONE %q, falling back to Local", cfg.CleanupTimezone)
		loc = time.Local
	}
	go scheduleDailyCleanup(cleanupCtx, svc, cleanupHour, cleanupMinute, loc)

	// 6. Start Server
	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port)
		if err := app.Listen(addr); err != nil {
			log.Printf("Server shutdown: %v", err)
		}
	}()

	// 7. Graceful Shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Gracefully shutting down...")
	cleanupCancel()
	if err := app.Shutdown(); err != nil {
		log.Printf("Error shutting down: %v", err)
	}
}

func scheduleDailyCleanup(ctx context.Context, svc *service.PriceService, hour, minute int, loc *time.Location) {
	for {
		now := time.Now().In(loc)
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc)
		if !next.After(now) {
			next = next.Add(24 * time.Hour)
		}

		timer := time.NewTimer(time.Until(next))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			removedTime, removedPrice, err := svc.ClearAllPrices(cleanupCtx)
			cancel()
			if err != nil {
				log.Printf("Cleanup failed: %v", err)
			} else {
				log.Printf("Cleanup finished: removed %d time records, %d price records", removedTime, removedPrice)
			}
		}
	}
}

func parseCleanupTime(value string) (int, int, error) {
	parsed, err := time.Parse("15:04", value)
	if err != nil {
		return 0, 0, err
	}
	return parsed.Hour(), parsed.Minute(), nil
}
