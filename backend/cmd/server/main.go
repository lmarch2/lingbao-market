package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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
		KeyGenerator: func(c *fiber.Ctx) string {
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
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// 5. Routes
	h := api.NewHandler(svc, authSvc, cfg.JWTSecret)
	h.RegisterRoutes(app)

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
	if err := app.Shutdown(); err != nil {
		log.Printf("Error shutting down: %v", err)
	}
}
