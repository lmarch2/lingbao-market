package api

import (
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lingbao-market/backend/internal/model"
	"github.com/lingbao-market/backend/internal/service"
)

type Handler struct {
	svc     *service.PriceService
	authSvc *service.AuthService
	jwtKey  []byte
}

func NewHandler(svc *service.PriceService, authSvc *service.AuthService, jwtKey string) *Handler {
	return &Handler{
		svc:     svc,
		authSvc: authSvc,
		jwtKey:  []byte(jwtKey),
	}
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	// Public
	api.Get("/feed", h.GetFeed)
	api.Get("/auth/captcha", h.GetCaptcha)
	api.Post("/auth/register", h.Register)
	api.Post("/auth/login", h.Login)

	// Protected
	api.Post("/submit", h.authMiddleware, h.SubmitPrice)
}

func (h *Handler) Register(c *fiber.Ctx) error {
	var req model.AuthRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	req.Username = strings.TrimSpace(req.Username)
	req.CaptchaID = strings.TrimSpace(req.CaptchaID)
	req.CaptchaCode = strings.TrimSpace(req.CaptchaCode)
	if len(req.Username) < 3 || len(req.Password) < 6 {
		return c.Status(400).JSON(fiber.Map{"error": "username min 3 chars, password min 6 chars"})
	}
	if ok, err := h.authSvc.VerifyCaptcha(c.Context(), req.CaptchaID, req.CaptchaCode); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "captcha verification failed"})
	} else if !ok {
		return c.Status(400).JSON(fiber.Map{"error": "invalid captcha"})
	}

	user, err := h.authSvc.Register(c.Context(), req.Username, req.Password)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	_ = h.authSvc.DeleteCaptcha(c.Context(), req.CaptchaID)

	return c.Status(201).JSON(fiber.Map{"id": user.ID, "username": user.Username})
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var req model.AuthRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	req.Username = strings.TrimSpace(req.Username)
	req.CaptchaID = strings.TrimSpace(req.CaptchaID)
	req.CaptchaCode = strings.TrimSpace(req.CaptchaCode)
	if ok, err := h.authSvc.VerifyCaptcha(c.Context(), req.CaptchaID, req.CaptchaCode); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "captcha verification failed"})
	} else if !ok {
		return c.Status(400).JSON(fiber.Map{"error": "invalid captcha"})
	}

	resp, err := h.authSvc.Login(c.Context(), req.Username, req.Password)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid credentials"})
	}
	_ = h.authSvc.DeleteCaptcha(c.Context(), req.CaptchaID)

	return c.JSON(resp)
}

func (h *Handler) GetCaptcha(c *fiber.Ctx) error {
	id, code, err := h.authSvc.CreateCaptcha(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create captcha"})
	}
	return c.JSON(fiber.Map{"captchaId": id, "code": code})
}

func (h *Handler) GetFeed(c *fiber.Ctx) error {
	sortBy := c.Query("sort", "time") // Default to time
	prices, err := h.svc.GetLatestFeed(c.Context(), 50, sortBy)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch feed"})
	}
	return c.JSON(prices)
}

func (h *Handler) SubmitPrice(c *fiber.Ctx) error {
	// User ID from middleware (optional usage)
	// claims := c.Locals("user").(jwt.MapClaims)
	// userID := claims["sub"].(string)

	var req model.SubmitRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	if req.Code == "" || req.Price <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "invalid data"})
	}

	// Validate Code Format (Alphanumeric, 3-12 chars)
	if match, _ := regexp.MatchString(`^[A-Z0-9]{3,12}$`, req.Code); !match {
		return c.Status(400).JSON(fiber.Map{"error": "invalid code format"})
	}

	item := model.PriceItem{
		Code:   req.Code,
		Price:  req.Price,
		Server: req.Server,
	}

	if err := h.svc.AddPrice(c.Context(), item); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to submit"})
	}

	return c.Status(201).JSON(fiber.Map{"status": "ok"})
}

func (h *Handler) authMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(401).JSON(fiber.Map{"error": "missing authorization header"})
	}

	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return h.jwtKey, nil
	})

	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{"error": "invalid token"})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "invalid token claims"})
	}

	c.Locals("user", claims)
	return c.Next()
}
